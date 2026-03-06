package worker

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/utils"
)

type PlannedStep struct {
	Type        domain.StepType
	DependsOn   []domain.StepType
	MaxAttempts int
	Payload     map[string]any
}

type JobPlan struct {
	Type     domain.JobType
	Priority domain.JobPriority
	Scope    domain.ScopePayload
	Steps    []PlannedStep
}

type JobPlanner interface {
	BuildPlan(jobType domain.JobType, scope domain.ScopePayload) (JobPlan, error)
}

type DefaultJobPlanner struct{}

func NewDefaultJobPlanner() *DefaultJobPlanner {
	return &DefaultJobPlanner{}
}

func (p *DefaultJobPlanner) BuildPlan(jobType domain.JobType, scope domain.ScopePayload) (JobPlan, error) {
	maxAttempts := defaultPlannedStepMaxAttempts()

	plan := JobPlan{
		Type:     jobType,
		Priority: domain.JobPriorityNormal,
		Scope:    scope,
	}

	switch jobType {
	case domain.JobTypeStartupScan, domain.JobTypeFSEvent, domain.JobTypeReindexFolder:
		plan.Steps = []PlannedStep{
			{Type: domain.StepTypeScanFilesystem, MaxAttempts: maxAttempts},
			{Type: domain.StepTypeDiffAgainstDB, DependsOn: []domain.StepType{domain.StepTypeScanFilesystem}, MaxAttempts: maxAttempts},
			{Type: domain.StepTypeMarkDeleted, DependsOn: []domain.StepType{domain.StepTypeDiffAgainstDB}, MaxAttempts: maxAttempts},
		}
	case domain.JobTypeUploadProcess:
		plan.Steps = []PlannedStep{
			{Type: domain.StepTypeMetadata, MaxAttempts: maxAttempts},
			{Type: domain.StepTypeChecksum, DependsOn: []domain.StepType{domain.StepTypeMetadata}, MaxAttempts: maxAttempts},
			{Type: domain.StepTypePersist, DependsOn: []domain.StepType{domain.StepTypeChecksum}, MaxAttempts: maxAttempts},
			{Type: domain.StepTypeThumbnail, DependsOn: []domain.StepType{domain.StepTypePersist}, MaxAttempts: maxAttempts},
			{Type: domain.StepTypePlaylistIndex, DependsOn: []domain.StepType{domain.StepTypePersist}, MaxAttempts: maxAttempts},
		}
	default:
		return JobPlan{}, fmt.Errorf("unsupported job type: %s", jobType)
	}

	return plan, nil
}

func defaultPlannedStepMaxAttempts() int {
	if config.AppConfig.WorkerRetryDefaultMaxAttempts > 0 {
		return config.AppConfig.WorkerRetryDefaultMaxAttempts
	}
	return 3
}

func buildFileProcessingPlan(filePath string, priority domain.JobPriority) JobPlan {
	fileFormatType := utils.GetFormatTypeByExtension(filepath.Ext(filePath)).Type
	maxAttempts := defaultPlannedStepMaxAttempts()

	steps := make([]PlannedStep, 0, 5)
	checksumDependencies := []domain.StepType{}

	if shouldExtractMetadata(fileFormatType) {
		steps = append(steps, PlannedStep{
			Type:        domain.StepTypeMetadata,
			MaxAttempts: maxAttempts,
		})
		checksumDependencies = append(checksumDependencies, domain.StepTypeMetadata)
	}

	steps = append(steps, PlannedStep{
		Type:        domain.StepTypeChecksum,
		DependsOn:   checksumDependencies,
		MaxAttempts: maxAttempts,
	})
	steps = append(steps, PlannedStep{
		Type:        domain.StepTypePersist,
		DependsOn:   []domain.StepType{domain.StepTypeChecksum},
		MaxAttempts: maxAttempts,
	})

	if shouldGenerateThumbnail(fileFormatType) {
		steps = append(steps, PlannedStep{
			Type:        domain.StepTypeThumbnail,
			DependsOn:   []domain.StepType{domain.StepTypePersist},
			MaxAttempts: maxAttempts,
		})
	}

	if fileFormatType == utils.FormatTypeVideo {
		steps = append(steps, PlannedStep{
			Type:        domain.StepTypePlaylistIndex,
			DependsOn:   []domain.StepType{domain.StepTypePersist},
			MaxAttempts: maxAttempts,
		})
	}

	return JobPlan{
		Type:     domain.JobTypeFSEvent,
		Priority: priority,
		Scope: domain.NewFileScopePayload(domain.FileScope{
			Name: filepath.Base(filePath),
			Path: filePath,
		}),
		Steps: steps,
	}
}

func shouldExtractMetadata(fileFormatType string) bool {
	return fileFormatType == utils.FormatTypeImage ||
		fileFormatType == utils.FormatTypeAudio ||
		fileFormatType == utils.FormatTypeVideo
}

func shouldGenerateThumbnail(fileFormatType string) bool {
	return fileFormatType == utils.FormatTypeImage ||
		fileFormatType == utils.FormatTypeVideo
}

type JobOrchestrator struct {
	repository jobs.RepositoryInterface
	planner    JobPlanner
	runInTx    func(fn func(*sql.Tx) error) error
}

func NewJobOrchestrator(repository jobs.RepositoryInterface, planner JobPlanner) *JobOrchestrator {
	if planner == nil {
		planner = NewDefaultJobPlanner()
	}

	orchestrator := &JobOrchestrator{
		repository: repository,
		planner:    planner,
	}

	if repository != nil {
		orchestrator.runInTx = func(fn func(*sql.Tx) error) error {
			dbContext := repository.GetDbContext()
			if dbContext == nil {
				return fmt.Errorf("jobs db context is nil")
			}
			return dbContext.ExecTx(fn)
		}
	}

	return orchestrator
}

func (o *JobOrchestrator) CreateJob(jobType domain.JobType, priority domain.JobPriority, scope domain.ScopePayload) (domain.Job, error) {
	plan, err := o.planner.BuildPlan(jobType, scope)
	if err != nil {
		return domain.Job{}, err
	}

	plan.Priority = priority
	if plan.Priority == 0 {
		plan.Priority = domain.JobPriorityNormal
	}

	return o.CreatePlannedJob(plan)
}

func (o *JobOrchestrator) CreatePlannedJob(plan JobPlan) (domain.Job, error) {
	if o == nil || o.repository == nil || o.runInTx == nil {
		return domain.Job{}, fmt.Errorf("job orchestrator is not configured")
	}

	jobID, err := newWorkerEntityID()
	if err != nil {
		return domain.Job{}, err
	}

	scopeJSON, err := marshalJSON(plan.Scope)
	if err != nil {
		return domain.Job{}, fmt.Errorf("marshal job scope: %w", err)
	}

	stepIDs := map[domain.StepType]string{}
	for _, step := range plan.Steps {
		if _, exists := stepIDs[step.Type]; exists {
			return domain.Job{}, fmt.Errorf("duplicated step type in plan: %s", step.Type)
		}

		stepID, stepIDErr := newWorkerEntityID()
		if stepIDErr != nil {
			return domain.Job{}, stepIDErr
		}
		stepIDs[step.Type] = stepID
	}

	err = o.runInTx(func(tx *sql.Tx) error {
		_, createErr := o.repository.CreateJob(tx, jobs.JobModel{
			ID:              jobID,
			Type:            string(plan.Type),
			Priority:        int(plan.Priority),
			ScopeJSON:       scopeJSON,
			Status:          string(domain.JobStatusQueued),
			CancelRequested: false,
			LastError:       "",
		})
		if createErr != nil {
			return createErr
		}

		for _, step := range plan.Steps {
			dependsOnIDs := make([]string, 0, len(step.DependsOn))
			for _, depType := range step.DependsOn {
				depID, exists := stepIDs[depType]
				if !exists {
					return fmt.Errorf("step %s depends on unknown step type %s", step.Type, depType)
				}
				dependsOnIDs = append(dependsOnIDs, depID)
			}

			dependsOnJSON, depErr := marshalJSON(dependsOnIDs)
			if depErr != nil {
				return fmt.Errorf("marshal step dependencies: %w", depErr)
			}

			payloadJSON, payloadErr := marshalJSON(step.Payload)
			if payloadErr != nil {
				return fmt.Errorf("marshal step payload: %w", payloadErr)
			}

			maxAttempts := step.MaxAttempts
			if maxAttempts <= 0 {
				maxAttempts = 1
			}

			_, createStepErr := o.repository.CreateStep(tx, jobs.StepModel{
				ID:            stepIDs[step.Type],
				JobID:         jobID,
				Type:          string(step.Type),
				Status:        string(domain.StepStatusQueued),
				DependsOnJSON: dependsOnJSON,
				Attempts:      0,
				MaxAttempts:   maxAttempts,
				LastError:     "",
				Progress:      0,
				PayloadJSON:   payloadJSON,
			})
			if createStepErr != nil {
				return createStepErr
			}
		}

		return nil
	})
	if err != nil {
		return domain.Job{}, fmt.Errorf("create job plan: %w", err)
	}

	now := time.Now().UTC()
	return domain.Job{
		ID:         jobID,
		Type:       plan.Type,
		Priority:   plan.Priority,
		Scope:      plan.Scope,
		Status:     domain.JobStatusQueued,
		CreatedAt:  now,
		UpdatedAt:  now,
		StartedAt:  nil,
		FinishedAt: nil,
	}, nil
}

type StepAtomicExecutor interface {
	ExecuteStep(step domain.Step, context *WorkerContext) error
}

func NewDefaultStepExecutor() *DefaultStepExecutor {
	return &DefaultStepExecutor{
		filesystemSnapshotsByJobID: map[string]map[string]files.FileDto{},
	}
}

type DefaultStepExecutor struct {
	snapshotMutex              sync.Mutex
	filesystemSnapshotsByJobID map[string]map[string]files.FileDto
}

func (e *DefaultStepExecutor) ExecuteStep(step domain.Step, context *WorkerContext) error {
	if context == nil {
		return fmt.Errorf("worker context is nil")
	}

	switch step.Type {
	case domain.StepTypeScanFilesystem:
		return e.executeScanFilesystemStep(step, context)
	case domain.StepTypeMetadata:
		fileInput, resolveErr := resolveStepFileInput(step, context)
		if resolveErr != nil {
			return resolveErr
		}
		_, execErr := NewMetadataStepExecutor(pythonScriptRunner).Execute(MetadataStepInput{File: fileInput})
		return execErr
	case domain.StepTypeChecksum:
		fileInput, resolveErr := resolveStepFileInput(step, context)
		if resolveErr != nil {
			return resolveErr
		}
		_, execErr := NewChecksumStepExecutor(utils.GetFileChecksum, utils.GetDirectoryChecksum).Execute(ChecksumStepInput{File: fileInput})
		return execErr
	case domain.StepTypePersist:
		fileInput, resolveErr := resolveStepFileInput(step, context)
		if resolveErr != nil {
			return resolveErr
		}
		_, execErr := NewPersistStepExecutor(context.FilesService).Execute(PersistStepInput{File: fileInput})
		return execErr
	case domain.StepTypeThumbnail:
		fileInput, resolveErr := resolveStepFileInput(step, context)
		if resolveErr != nil {
			return resolveErr
		}
		_, execErr := NewThumbnailStepExecutor(context.FilesService).Execute(ThumbnailStepInput{File: &fileInput})
		return execErr
	case domain.StepTypePlaylistIndex:
		_, execErr := NewPlaylistIndexStepExecutor(context.VideoService).Execute(PlaylistIndexStepInput{})
		return execErr
	case domain.StepTypeDiffAgainstDB:
		return e.executeDiffAgainstDBStep(step, context)
	case domain.StepTypeMarkDeleted:
		return e.executeMarkDeletedStep(step, context)
	default:
		return fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

func (e *DefaultStepExecutor) executeScanFilesystemStep(step domain.Step, context *WorkerContext) error {
	rootPath := resolveScopeRoot(step.Scope)
	if rootPath == "" {
		return fmt.Errorf("scan_filesystem step: root path is required")
	}

	shouldContinue := buildJobCancellationChecker(context, step.JobID)
	snapshot, err := collectFilesystemSnapshot(rootPath, shouldContinue)
	if err != nil {
		return err
	}

	e.storeSnapshot(step.JobID, snapshot)
	return nil
}

func (e *DefaultStepExecutor) executeDiffAgainstDBStep(step domain.Step, context *WorkerContext) (executionErr error) {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("diff_against_db step: file service is required")
	}
	if context.Orchestrator == nil {
		return fmt.Errorf("diff_against_db step: job orchestrator is required")
	}

	defer func() {
		if executionErr != nil {
			e.clearSnapshot(step.JobID)
		}
	}()

	snapshot := e.loadSnapshot(step.JobID)
	if len(snapshot) == 0 {
		rootPath := resolveScopeRoot(step.Scope)
		if rootPath == "" {
			return fmt.Errorf("diff_against_db step: root path is required")
		}

		var scanErr error
		shouldContinue := buildJobCancellationChecker(context, step.JobID)
		snapshot, scanErr = collectFilesystemSnapshot(rootPath, shouldContinue)
		if scanErr != nil {
			return scanErr
		}
	}

	shouldContinue := buildJobCancellationChecker(context, step.JobID)
	dbFiles, err := listAllFiles(context.FilesService, shouldContinue)
	if err != nil {
		return err
	}

	dbByPath := make(map[string]files.FileDto, len(dbFiles))
	for _, dbFile := range dbFiles {
		dbByPath[dbFile.Path] = dbFile
	}

	entriesToProcess := make([]files.FileDto, 0, len(snapshot))
	newCount := 0
	modifiedCount := 0
	unchangedCount := 0
	reactivatedCount := 0

	for path, currentEntry := range snapshot {
		existingEntry, exists := dbByPath[path]
		if !exists {
			entriesToProcess = append(entriesToProcess, currentEntry)
			newCount++
			continue
		}

		delete(dbByPath, path)
		if !existingEntry.DeletedAt.HasValue && isEntryUnchanged(existingEntry, currentEntry) {
			unchangedCount++
			continue
		}

		currentEntry.ID = existingEntry.ID
		currentEntry.CheckSum = existingEntry.CheckSum
		currentEntry.Starred = existingEntry.Starred
		currentEntry.LastInteraction = existingEntry.LastInteraction
		currentEntry.LastBackup = existingEntry.LastBackup
		currentEntry.Metadata = nil

		entriesToProcess = append(entriesToProcess, currentEntry)
		if existingEntry.DeletedAt.HasValue {
			reactivatedCount++
		} else {
			modifiedCount++
		}
	}

	if err := enqueueProcessEntries(context, step, entriesToProcess, shouldContinue); err != nil {
		return err
	}

	log.Printf(
		"startup_scan diff concluido job_id=%s new=%d modified=%d unchanged=%d reactivated=%d",
		step.JobID,
		newCount,
		modifiedCount,
		unchangedCount,
		reactivatedCount,
	)

	return nil
}

func (e *DefaultStepExecutor) executeMarkDeletedStep(step domain.Step, context *WorkerContext) error {
	if context == nil || context.FilesService == nil {
		return fmt.Errorf("mark_deleted step: file service is required")
	}

	snapshot := e.loadSnapshot(step.JobID)
	if len(snapshot) == 0 {
		rootPath := resolveScopeRoot(step.Scope)
		if rootPath == "" {
			return fmt.Errorf("mark_deleted step: root path is required")
		}

		var scanErr error
		shouldContinue := buildJobCancellationChecker(context, step.JobID)
		snapshot, scanErr = collectFilesystemSnapshot(rootPath, shouldContinue)
		if scanErr != nil {
			return scanErr
		}
	}
	defer e.clearSnapshot(step.JobID)

	shouldContinue := buildJobCancellationChecker(context, step.JobID)
	rootScopePath := resolveScopeRoot(step.Scope)
	if rootScopePath == "" {
		return fmt.Errorf("mark_deleted step: root path is required")
	}

	activeDBFiles, err := listActiveFiles(context.FilesService, rootScopePath, shouldContinue)
	if err != nil {
		return err
	}

	deletedEntries := make([]files.FileDto, 0)
	for _, activeEntry := range activeDBFiles {
		if _, exists := snapshot[activeEntry.Path]; !exists {
			deletedEntries = append(deletedEntries, activeEntry)
		}
	}

	if err := markDeletedEntries(context.FilesService, deletedEntries, shouldContinue); err != nil {
		return err
	}

	log.Printf(
		"mark_deleted concluido job_id=%s deleted=%d",
		step.JobID,
		len(deletedEntries),
	)

	return nil
}

func (e *DefaultStepExecutor) storeSnapshot(jobID string, snapshot map[string]files.FileDto) {
	if e == nil || jobID == "" {
		return
	}

	e.snapshotMutex.Lock()
	defer e.snapshotMutex.Unlock()
	e.filesystemSnapshotsByJobID[jobID] = snapshot
}

func (e *DefaultStepExecutor) loadSnapshot(jobID string) map[string]files.FileDto {
	if e == nil || jobID == "" {
		return nil
	}

	e.snapshotMutex.Lock()
	defer e.snapshotMutex.Unlock()

	snapshot := e.filesystemSnapshotsByJobID[jobID]
	if len(snapshot) == 0 {
		return nil
	}

	clone := make(map[string]files.FileDto, len(snapshot))
	for path, dto := range snapshot {
		clone[path] = dto
	}
	return clone
}

func (e *DefaultStepExecutor) clearSnapshot(jobID string) {
	if e == nil || jobID == "" {
		return
	}

	e.snapshotMutex.Lock()
	defer e.snapshotMutex.Unlock()
	delete(e.filesystemSnapshotsByJobID, jobID)
}

func resolveScopeRoot(scope domain.ScopePayload) string {
	if scope.Root != nil && scope.Root.Root != "" {
		return scope.Root.Root
	}
	if scope.Path != nil && scope.Path.Path != "" {
		return scope.Path.Path
	}
	return ""
}

func collectFilesystemSnapshot(rootPath string, shouldContinue func() error) (map[string]files.FileDto, error) {
	entries := map[string]files.FileDto{}

	walkErr := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if shouldContinue != nil {
			if cancelErr := shouldContinue(); cancelErr != nil {
				return cancelErr
			}
		}

		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return nil
			}
			return err
		}

		fileEntry := files.FileDto{
			Path:       path,
			ParentPath: filepath.Dir(path),
		}
		if parseErr := fileEntry.ParseFileInfoToFileDto(info); parseErr != nil {
			return parseErr
		}
		fileEntry.DeletedAt = utils.Optional[time.Time]{HasValue: false}
		fileEntry.Metadata = nil

		entries[path] = fileEntry
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("scan_filesystem step: %w", walkErr)
	}

	return entries, nil
}

func listActiveFiles(service files.ServiceInterface, rootPath string, shouldContinue func() error) ([]files.FileDto, error) {
	allFiles, err := listAllFiles(service, shouldContinue)
	if err != nil {
		return nil, err
	}

	cleanRoot := filepath.Clean(rootPath)
	activeEntries := make([]files.FileDto, 0, len(allFiles))
	for _, fileEntry := range allFiles {
		if fileEntry.DeletedAt.HasValue {
			continue
		}
		if !isPathWithinScope(fileEntry.Path, cleanRoot) {
			continue
		}
		activeEntries = append(activeEntries, fileEntry)
	}
	return activeEntries, nil
}

func listAllFiles(service files.ServiceInterface, shouldContinue func() error) ([]files.FileDto, error) {
	allEntries := make([]files.FileDto, 0)
	currentPage := 1
	const pageSize = 500

	for {
		if shouldContinue != nil {
			if cancelErr := shouldContinue(); cancelErr != nil {
				return nil, cancelErr
			}
		}

		pageResult, err := service.GetFiles(files.FileFilter{}, currentPage, pageSize)
		if err != nil {
			return nil, fmt.Errorf("diff_against_db step: list db files: %w", err)
		}

		allEntries = append(allEntries, pageResult.Items...)

		if !pageResult.Pagination.HasNext {
			break
		}
		currentPage++
	}

	return allEntries, nil
}

func isEntryUnchanged(dbEntry files.FileDto, filesystemEntry files.FileDto) bool {
	return dbEntry.Size == filesystemEntry.Size && dbEntry.UpdatedAt.Equal(filesystemEntry.UpdatedAt)
}

func isPathWithinScope(filePath string, scopePath string) bool {
	cleanFilePath := filepath.Clean(filePath)
	cleanScopePath := filepath.Clean(scopePath)

	if cleanFilePath == cleanScopePath {
		return true
	}
	return strings.HasPrefix(cleanFilePath, cleanScopePath+string(filepath.Separator))
}

func markDeletedEntries(service files.ServiceInterface, deletedEntries []files.FileDto, shouldContinue func() error) error {
	if len(deletedEntries) == 0 {
		return nil
	}

	now := time.Now()
	for _, deletedEntry := range deletedEntries {
		if shouldContinue != nil {
			if cancelErr := shouldContinue(); cancelErr != nil {
				return cancelErr
			}
		}

		deletedEntry.DeletedAt = utils.Optional[time.Time]{
			HasValue: true,
			Value:    now,
		}

		updated, err := service.UpdateFile(deletedEntry)
		if err != nil {
			return fmt.Errorf("diff_against_db step: mark deleted file %d: %w", deletedEntry.ID, err)
		}
		if !updated {
			return fmt.Errorf("diff_against_db step: file %d was not marked as deleted", deletedEntry.ID)
		}
	}

	return nil
}

func enqueueProcessEntries(context *WorkerContext, step domain.Step, entries []files.FileDto, shouldContinue func() error) error {
	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if shouldContinue != nil {
			if cancelErr := shouldContinue(); cancelErr != nil {
				return cancelErr
			}
		}

		plan := JobPlan{
			Type:     domain.JobTypeFSEvent,
			Priority: domain.JobPriorityLow,
			Scope: domain.NewFileScopePayload(domain.FileScope{
				ID:   entry.ID,
				Name: entry.Name,
				Path: entry.Path,
			}),
			Steps: []PlannedStep{
				{
					Type:        domain.StepTypePersist,
					MaxAttempts: defaultPlannedStepMaxAttempts(),
				},
			},
		}
		if entry.Type == files.File {
			plan = buildFileProcessingPlan(entry.Path, domain.JobPriorityLow)
			plan.Scope = domain.NewFileScopePayload(domain.FileScope{
				ID:   entry.ID,
				Name: entry.Name,
				Path: entry.Path,
			})
		}
		job, err := context.Orchestrator.CreatePlannedJob(plan)
		if err != nil {
			return fmt.Errorf("diff_against_db fan-out enqueue file=%s: %w", entry.Path, err)
		}
		log.Printf("diff_against_db enfileirou processamento incremental parent_job_id=%s child_job_id=%s path=%s", step.JobID, job.ID, entry.Path)
	}

	return nil
}

func buildJobCancellationChecker(workerContext *WorkerContext, jobID string) func() error {
	return func() error {
		if workerContext == nil || jobID == "" {
			return nil
		}
		return workerContext.CheckJobCancellation(jobID)
	}
}

func resolveStepFileInput(step domain.Step, context *WorkerContext) (files.FileDto, error) {
	if context == nil || context.FilesService == nil {
		return files.FileDto{}, fmt.Errorf("files service is required")
	}

	if step.Scope.File != nil {
		if step.Scope.File.ID > 0 {
			return context.FilesService.GetFileById(step.Scope.File.ID)
		}

		if step.Scope.File.Path != "" {
			fileInfo, statErr := os.Stat(step.Scope.File.Path)
			if statErr != nil {
				return files.FileDto{}, statErr
			}

			file := files.FileDto{
				ID:   step.Scope.File.ID,
				Name: step.Scope.File.Name,
				Path: step.Scope.File.Path,
			}

			if file.Name == "" {
				file.Name = filepath.Base(step.Scope.File.Path)
			}
			file.ParentPath = filepath.Dir(step.Scope.File.Path)
			if parseErr := file.ParseFileInfoToFileDto(fileInfo); parseErr != nil {
				return files.FileDto{}, parseErr
			}
			file.Path = step.Scope.File.Path
			return file, nil
		}
	}

	return files.FileDto{}, fmt.Errorf("step %s requires file scope", step.Type)
}

func marshalJSON(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

func newWorkerEntityID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
