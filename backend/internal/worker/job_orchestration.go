package worker

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
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
	plan := JobPlan{
		Type:     jobType,
		Priority: domain.JobPriorityNormal,
		Scope:    scope,
	}

	switch jobType {
	case domain.JobTypeStartupScan, domain.JobTypeFSEvent, domain.JobTypeReindexFolder:
		plan.Steps = []PlannedStep{
			{Type: domain.StepTypeScanFilesystem, MaxAttempts: 1},
			{Type: domain.StepTypeDiffAgainstDB, DependsOn: []domain.StepType{domain.StepTypeScanFilesystem}, MaxAttempts: 1},
		}
	case domain.JobTypeUploadProcess:
		plan.Steps = []PlannedStep{
			{Type: domain.StepTypeMetadata, MaxAttempts: 1},
			{Type: domain.StepTypeChecksum, DependsOn: []domain.StepType{domain.StepTypeMetadata}, MaxAttempts: 1},
			{Type: domain.StepTypePersist, DependsOn: []domain.StepType{domain.StepTypeChecksum}, MaxAttempts: 1},
			{Type: domain.StepTypeThumbnail, DependsOn: []domain.StepType{domain.StepTypePersist}, MaxAttempts: 1},
			{Type: domain.StepTypePlaylistIndex, DependsOn: []domain.StepType{domain.StepTypePersist}, MaxAttempts: 1},
		}
	default:
		return JobPlan{}, fmt.Errorf("unsupported job type: %s", jobType)
	}

	return plan, nil
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

type DefaultStepExecutor struct{}

func NewDefaultStepExecutor() *DefaultStepExecutor {
	return &DefaultStepExecutor{}
}

func (e *DefaultStepExecutor) ExecuteStep(step domain.Step, context *WorkerContext) error {
	if context == nil {
		return fmt.Errorf("worker context is nil")
	}

	switch step.Type {
	case domain.StepTypeScanFilesystem:
		ScanFilesWorker(context.FilesService, context.Logger)
		return nil
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
	case domain.StepTypeDiffAgainstDB,
		domain.StepTypeMarkDeleted:
		// Placeholder during migration: scheduler executes one atomic step at a time.
		return nil
	default:
		return fmt.Errorf("unsupported step type: %s", step.Type)
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
