package engine

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	jobdomain "nas-go/api/internal/worker/job"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

type fakeWorkerNotifSvc struct {
	notifications.ServiceInterface
	dtos []notifications.CreateNotificationDto
}

func (f *fakeWorkerNotifSvc) GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error) {
	f.dtos = append(f.dtos, dto)
	return notifications.NotificationDto{}, nil
}

type fakeEngineImageRepository struct {
	imagedom.RepositoryInterface
	dbCtx    *database.DbContext
	mock     sqlmock.Sqlmock
	upsertFn func(tx *sql.Tx, m imagedom.MetadataModel) (imagedom.MetadataModel, error)
}

func newFakeEngineImageRepository(upsertFn func(tx *sql.Tx, m imagedom.MetadataModel) (imagedom.MetadataModel, error)) *fakeEngineImageRepository {
	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	mock.ExpectCommit()
	return &fakeEngineImageRepository{
		dbCtx:    database.NewDbContext(db),
		mock:     mock,
		upsertFn: upsertFn,
	}
}

func (f *fakeEngineImageRepository) GetDbContext() *database.DbContext {
	return f.dbCtx
}

func (f *fakeEngineImageRepository) UpsertImageMetadata(tx *sql.Tx, m imagedom.MetadataModel) (imagedom.MetadataModel, error) {
	if f.upsertFn != nil {
		return f.upsertFn(tx, m)
	}
	return m, nil
}

func TestBuildStepExecutorsAndPlans(t *testing.T) {
	executors := buildStepExecutors(&WorkerContext{})
	if len(executors) != 12 {
		t.Fatalf("expected 12 step executors, got %d", len(executors))
	}

	imagePlan, err := buildFileProcessingPlan(
		files.FileDto{Path: "/tmp/image.jpg", Format: ".jpg"},
		jobdomain.JobTypeFSEvent,
		jobdomain.JobPriorityLow,
	)
	if err != nil {
		t.Fatalf("buildFileProcessingPlan image error: %v", err)
	}
	if len(imagePlan.Steps) != 4 || imagePlan.Steps[3].Type != jobdomain.StepTypeThumbnail {
		t.Fatalf("unexpected image plan: %+v", imagePlan)
	}

	videoPlan, err := buildFileProcessingPlan(
		files.FileDto{Path: "/tmp/video.mp4", Format: ".mp4"},
		jobdomain.JobTypeFSEvent,
		jobdomain.JobPriorityLow,
	)
	if err != nil {
		t.Fatalf("buildFileProcessingPlan video error: %v", err)
	}
	if len(videoPlan.Steps) != 5 || videoPlan.Steps[4].Type != jobdomain.StepTypePlaylistIndex {
		t.Fatalf("unexpected video plan: %+v", videoPlan)
	}

	scanPlan, err := buildScanPlan("/data", jobdomain.JobTypeStartupScan, jobdomain.JobPriorityNormal)
	if err != nil {
		t.Fatalf("buildScanPlan error: %v", err)
	}
	if len(scanPlan.Steps) != 3 || scanPlan.Steps[0].Type != jobdomain.StepTypeScanFilesystem || scanPlan.Steps[1].Type != jobdomain.StepTypeDiffAgainstDB || scanPlan.Steps[2].Type != jobdomain.StepTypeMarkDeleted {
		t.Fatalf("unexpected scan plan: %+v", scanPlan)
	}

	payload, err := marshalPayload(StepFilePayload{Path: "/tmp/path"})
	if err != nil || len(payload) == 0 {
		t.Fatalf("marshalPayload returned payload=%q err=%v", string(payload), err)
	}
}

func TestExecuteScanFilesystemStepBranches(t *testing.T) {
	if err := executeScanFilesystemStep(nil, jobs.StepModel{}); err == nil {
		t.Fatalf("expected nil context error")
	}

	if err := executeScanFilesystemStep(&WorkerContext{}, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected empty payload skip, got %v", err)
	}

	missingPayload, _ := marshalPayload(StepFilePayload{Path: filepath.Join(t.TempDir(), "missing")})
	if err := executeScanFilesystemStep(&WorkerContext{}, jobs.StepModel{Payload: missingPayload}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected missing path skip, got %v", err)
	}

	root := t.TempDir()
	filePath := filepath.Join(root, "item.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	filePayload, _ := marshalPayload(StepFilePayload{Path: filePath})
	if err := executeScanFilesystemStep(&WorkerContext{}, jobs.StepModel{Payload: filePayload}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected file path skip, got %v", err)
	}

	dirPayload, _ := marshalPayload(StepFilePayload{Path: root})
	if err := executeScanFilesystemStep(&WorkerContext{}, jobs.StepModel{Payload: dirPayload}); err != nil {
		t.Fatalf("expected directory scan success, got %v", err)
	}
}

func TestResolveFileDtoForStepAndMetadataStepSuccess(t *testing.T) {
	inlineFile := files.FileDto{ID: 1, Name: "inline.jpg", Path: "/tmp/inline.jpg", ParentPath: "/tmp", Format: ".jpg", Type: files.File}
	resolved, err := resolveFileDtoForStep(&workerFilesServiceMock{}, StepFilePayload{File: &inlineFile})
	if err != nil || resolved.ID != inlineFile.ID {
		t.Fatalf("resolveFileDtoForStep inline returned %+v err=%v", resolved, err)
	}

	if _, err := resolveFileDtoForStep(
		&workerFilesServiceMock{
			getFileByIDFn: func(id int) (files.FileDto, error) { return files.FileDto{}, sql.ErrNoRows },
		},
		StepFilePayload{FileID: 7},
	); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected skipped file-id lookup, got %v", err)
	}

	root := t.TempDir()
	filePath := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	parsed, err := resolveFileDtoForStep(
		&workerFilesServiceMock{
			getFileByNamePathFn: func(name, path string) (files.FileDto, error) { return files.FileDto{}, sql.ErrNoRows },
		},
		StepFilePayload{Path: filePath},
	)
	if err != nil || parsed.Path != filePath || parsed.Name != "photo.jpg" {
		t.Fatalf("resolveFileDtoForStep path returned %+v err=%v", parsed, err)
	}

	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(imagedom.MetadataModel{Path: filePath, Format: "jpg"})
		return string(payload), nil
	})
	defer scan.SetPythonScriptRunnerForTesting(nil)

	upserted := 0
	fakeImageRepo := newFakeEngineImageRepository(func(tx *sql.Tx, m imagedom.MetadataModel) (imagedom.MetadataModel, error) {
		upserted++
		return m, nil
	})
	filesService := &workerFilesServiceMock{}
	payload, _ := marshalPayload(StepFilePayload{
		File: &files.FileDto{
			ID:         10,
			Name:       "photo.jpg",
			Path:       filePath,
			ParentPath: root,
			Format:     ".jpg",
			Type:       files.File,
		},
	})
	if err := executeMetadataStep(&WorkerContext{FilesService: filesService, ImageRepository: fakeImageRepo}, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeMetadataStep returned error: %v", err)
	}
	if upserted != 1 {
		t.Fatalf("expected image metadata upsert, got %d upserts", upserted)
	}
}

func TestExecuteDiffAgainstDBStepIndexesDirectories(t *testing.T) {
	prevEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = prevEntryPoint })

	root := t.TempDir()
	config.AppConfig.EntryPoint = root

	knownDir := filepath.Join(root, "known")
	newDir := filepath.Join(root, "musicas")
	newNestedDir := filepath.Join(newDir, "album novo")
	for _, dir := range []string{knownDir, newNestedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll %s failed: %v", dir, err)
		}
	}

	created := []files.FileDto{}
	filesService := &workerFilesServiceMock{
		getFileStatByPathFn: func(path string) (files.FileStat, bool, error) {
			// Only knownDir already has an active row in the DB.
			return files.FileStat{}, path == knownDir, nil
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			created = append(created, fileDto)
			return fileDto, nil
		},
	}

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)

	diffPayload, _ := marshalPayload(StepFilePayload{Path: root})
	err := executeDiffAgainstDBStep(
		&WorkerContext{FilesService: filesService, JobOrchestrator: orchestrator},
		jobs.StepModel{Payload: diffPayload},
	)
	// Directories are upserted inline, not enqueued — with no files in the
	// tree the step reports nothing sent to the pipeline.
	if !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped for a files-free tree, got %v", err)
	}
	if len(repository.jobs) != 0 {
		t.Fatalf("directories must not enqueue processing jobs, got %d", len(repository.jobs))
	}

	// The scanned root gets a row too: storage roots are the level-zero
	// nodes of the multi-root tree.
	if len(created) != 3 {
		t.Fatalf("expected rows created for the root and the 2 missing directories, got %+v", created)
	}
	createdPaths := map[string]files.FileType{}
	for _, fileDto := range created {
		createdPaths[fileDto.Path] = fileDto.Type
	}
	for _, dir := range []string{root, newDir, newNestedDir} {
		if createdPaths[dir] != files.Directory {
			t.Fatalf("expected directory row for %q, got %+v", dir, createdPaths)
		}
	}
	if _, ok := createdPaths[knownDir]; ok {
		t.Fatalf("directory with an existing active row must not be recreated")
	}
}

func TestExecuteDiffAgainstDBStepAndMarkDeletedStep(t *testing.T) {
	root := t.TempDir()
	unchangedPath := filepath.Join(root, "same.txt")
	changedPath := filepath.Join(root, "new.txt")
	existingPath := filepath.Join(root, "exists.txt")

	for _, filePath := range []string{unchangedPath, changedPath, existingPath} {
		if err := os.WriteFile(filePath, []byte(filePath), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", filePath, err)
		}
	}

	unchangedInfo, err := os.Stat(unchangedPath)
	if err != nil {
		t.Fatalf("Stat unchanged file failed: %v", err)
	}

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	filesService := &workerFilesServiceMock{
		getFileStatByPathFn: func(path string) (files.FileStat, bool, error) {
			// Only the unchanged file is known to the DB with a matching
			// size + mtime; the other two paths are new, so they must be
			// enqueued for processing.
			if path == unchangedPath {
				return files.FileStat{
					Size:      unchangedInfo.Size(),
					UpdatedAt: unchangedInfo.ModTime(),
				}, true, nil
			}
			return files.FileStat{}, false, nil
		},
	}

	diffPayload, _ := marshalPayload(StepFilePayload{Path: root})
	notifSvc := &fakeWorkerNotifSvc{}
	err = executeDiffAgainstDBStep(
		&WorkerContext{FilesService: filesService, JobOrchestrator: orchestrator, NotificationService: notifSvc},
		jobs.StepModel{Payload: diffPayload},
	)
	if err != nil {
		t.Fatalf("executeDiffAgainstDBStep returned error: %v", err)
	}
	if len(repository.jobs) != 2 {
		t.Fatalf("expected two created jobs for changed files, got %d", len(repository.jobs))
	}
	// A scan that identified files to process emits a single completion
	// notification (info) reporting that files were enqueued.
	if len(notifSvc.dtos) != 1 {
		t.Fatalf("expected one scan-completed notification, got %d", len(notifSvc.dtos))
	}
	if notifSvc.dtos[0].Type != "info" {
		t.Fatalf("expected info notification, got %q", notifSvc.dtos[0].Type)
	}
	if notifSvc.dtos[0].Title != i18n.GetMessage("NOTIFICATION_FILE_SCAN_COMPLETED_TITLE") {
		t.Fatalf("unexpected notification title %q", notifSvc.dtos[0].Title)
	}

	// Running the scan again must NOT re-enqueue the files that already have a
	// pending job: idempotency skips them, so no new jobs are created and the
	// step reports nothing to process (ErrStepSkipped) instead of re-counting
	// every candidate. This is what keeps the pipeline from being flooded with
	// the same files on every startup.
	notifSvc.dtos = nil
	if err := executeDiffAgainstDBStep(
		&WorkerContext{FilesService: filesService, JobOrchestrator: orchestrator, NotificationService: notifSvc},
		jobs.StepModel{Payload: diffPayload},
	); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected second scan to skip (no new files), got %v", err)
	}
	if len(repository.jobs) != 2 {
		t.Fatalf("expected no additional jobs on second scan, got %d", len(repository.jobs))
	}
	if len(notifSvc.dtos) != 0 {
		t.Fatalf("expected no completion notification on second scan, got %d", len(notifSvc.dtos))
	}

	errBoom := errors.New("list failed")
	if err := executeDiffAgainstDBStep(
		&WorkerContext{
			FilesService: &workerFilesServiceMock{
				getFileStatByPathFn: func(path string) (files.FileStat, bool, error) {
					return files.FileStat{}, false, errBoom
				},
			},
			JobOrchestrator: orchestrator,
		},
		jobs.StepModel{Payload: diffPayload},
	); !errors.Is(err, errBoom) {
		t.Fatalf("expected diff step error, got %v", err)
	}

	updates := []files.FileDto{}
	missingPath := filepath.Join(root, "missing.txt")
	restoreTime := time.Now().Add(-time.Hour)
	markDeletedService := &workerFilesServiceMock{
		getByPathPrefixFn: func(prefix string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{ID: 10, Path: missingPath},
					{
						ID:        11,
						Path:      existingPath,
						DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: restoreTime},
					},
				},
				Pagination: utils.Pagination{Page: page, PageSize: pageSize},
			}, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updates = append(updates, file)
			return true, nil
		},
	}

	if err := executeMarkDeletedStep(&WorkerContext{FilesService: markDeletedService}, jobs.StepModel{Payload: diffPayload}); err != nil {
		t.Fatalf("executeMarkDeletedStep returned error: %v", err)
	}
	if len(updates) != 2 || !updates[0].DeletedAt.HasValue || updates[1].DeletedAt.HasValue {
		t.Fatalf("unexpected mark-deleted updates: %+v", updates)
	}

	noChangesService := &workerFilesServiceMock{
		getByPathPrefixFn: func(prefix string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items:      []files.FileDto{{ID: 12, Path: existingPath}},
				Pagination: utils.Pagination{Page: page, PageSize: pageSize},
			}, nil
		},
	}
	if err := executeMarkDeletedStep(&WorkerContext{FilesService: noChangesService}, jobs.StepModel{Payload: diffPayload}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected mark-deleted skip, got %v", err)
	}
}

func TestWorkerEnqueueHelpers(t *testing.T) {
	previousConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = previousConfig
	})

	root := t.TempDir()
	config.AppConfig.EntryPoint = root

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	if err := enqueueStartupScanJob(context); err != nil {
		t.Fatalf("enqueueStartupScanJob returned error: %v", err)
	}
	if len(repository.jobs) != 1 {
		t.Fatalf("expected startup scan job, got %d jobs", len(repository.jobs))
	}

	if err := enqueueFilesystemEventJob(context, root, jobdomain.JobPriorityHigh); err != nil {
		t.Fatalf("enqueueFilesystemEventJob returned error: %v", err)
	}
	if len(repository.jobs) != 2 {
		t.Fatalf("expected two queued jobs, got %d", len(repository.jobs))
	}

	if err := enqueueFilesystemEventJob(context, "", jobdomain.JobPriorityLow); err != nil {
		t.Fatalf("expected empty root enqueue to be ignored, got %v", err)
	}

	startupStepPayload, _ := marshalPayload(StepFilePayload{Path: root})
	if string(repository.steps[1].Payload) == "" || string(startupStepPayload) == "" {
		t.Fatalf("expected persisted step payloads")
	}

	if _, err := buildScanPlan("", jobdomain.JobTypeStartupScan, jobdomain.JobPriorityLow); err != nil {
		t.Fatalf("buildScanPlan should allow empty roots, got %v", err)
	}

	for jobID, job := range repository.jobs {
		if job.Type == "" || job.Priority == "" {
			t.Fatalf("job %d not populated: %+v", jobID, job)
		}
	}

	if _, err := repository.GetStepsByJobID(1); err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}

	for _, step := range repository.steps {
		if len(step.Payload) == 0 {
			t.Fatalf("expected non-empty step payload for step %+v", step)
		}
	}

	if err := executeDiffAgainstDBStep(nil, jobs.StepModel{}); err == nil || err.Error() != "files service and orchestrator are required for diff step" {
		t.Fatalf("expected diff-step context error, got %v", err)
	}

	if err := executeMarkDeletedStep(nil, jobs.StepModel{}); err == nil || err.Error() != "files service is required for mark_deleted step" {
		t.Fatalf("expected mark-deleted context error, got %v", err)
	}

	if err := executeDiffAgainstDBStep(
		&WorkerContext{FilesService: &workerFilesServiceMock{}, JobOrchestrator: orchestrator},
		jobs.StepModel{Payload: []byte(fmt.Sprintf("{invalid-%s", root))},
	); err == nil {
		t.Fatalf("expected diff-step payload decode error")
	}
}
