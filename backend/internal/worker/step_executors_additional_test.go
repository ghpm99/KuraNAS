package worker

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"
)

func TestBuildStepExecutorsAndPlans(t *testing.T) {
	executors := buildStepExecutors(&WorkerContext{})
	if len(executors) != 8 {
		t.Fatalf("expected 8 step executors, got %d", len(executors))
	}

	imagePlan, err := buildFileProcessingPlan(
		files.FileDto{Path: "/tmp/image.jpg", Format: ".jpg"},
		JobTypeFSEvent,
		JobPriorityLow,
	)
	if err != nil {
		t.Fatalf("buildFileProcessingPlan image error: %v", err)
	}
	if len(imagePlan.Steps) != 4 || imagePlan.Steps[3].Type != StepTypeThumbnail {
		t.Fatalf("unexpected image plan: %+v", imagePlan)
	}

	videoPlan, err := buildFileProcessingPlan(
		files.FileDto{Path: "/tmp/video.mp4", Format: ".mp4"},
		JobTypeFSEvent,
		JobPriorityLow,
	)
	if err != nil {
		t.Fatalf("buildFileProcessingPlan video error: %v", err)
	}
	if len(videoPlan.Steps) != 5 || videoPlan.Steps[4].Type != StepTypePlaylistIndex {
		t.Fatalf("unexpected video plan: %+v", videoPlan)
	}

	scanPlan, err := buildScanPlan("/data", JobTypeStartupScan, JobPriorityNormal)
	if err != nil {
		t.Fatalf("buildScanPlan error: %v", err)
	}
	if len(scanPlan.Steps) != 3 || scanPlan.Steps[0].Type != StepTypeScanFilesystem || scanPlan.Steps[1].Type != StepTypeDiffAgainstDB || scanPlan.Steps[2].Type != StepTypeMarkDeleted {
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

	SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(files.ImageMetadataModel{Path: filePath, Format: "jpg"})
		return string(payload), nil
	})
	defer SetPythonScriptRunnerForTesting(nil)

	updated := 0
	filesService := &workerFilesServiceMock{
		updateFileFn: func(file files.FileDto) (bool, error) {
			updated++
			if file.Metadata == nil {
				t.Fatalf("expected metadata on updated file")
			}
			return true, nil
		},
	}
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
	if err := executeMetadataStep(&WorkerContext{FilesService: filesService}, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeMetadataStep returned error: %v", err)
	}
	if updated != 1 {
		t.Fatalf("expected metadata update, got %d updates", updated)
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
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			switch filter.PathPrefix.Value {
			case root:
				return utils.PaginationResponse[files.FileDto]{
					Items: []files.FileDto{
						{
							ID:        1,
							Name:      "same.txt",
							Path:      unchangedPath,
							Size:      unchangedInfo.Size(),
							UpdatedAt: unchangedInfo.ModTime(),
						},
					},
					Pagination: utils.Pagination{Page: page, PageSize: pageSize},
				}, nil
			default:
				return utils.PaginationResponse[files.FileDto]{}, nil
			}
		},
	}

	diffPayload, _ := marshalPayload(StepFilePayload{Path: root})
	err = executeDiffAgainstDBStep(
		&WorkerContext{FilesService: filesService, JobOrchestrator: orchestrator},
		jobs.StepModel{Payload: diffPayload},
	)
	if err != nil {
		t.Fatalf("executeDiffAgainstDBStep returned error: %v", err)
	}
	if len(repository.jobs) != 2 {
		t.Fatalf("expected two created jobs for changed files, got %d", len(repository.jobs))
	}

	errBoom := errors.New("list failed")
	if err := executeDiffAgainstDBStep(
		&WorkerContext{
			FilesService: &workerFilesServiceMock{
				getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
					return utils.PaginationResponse[files.FileDto]{}, errBoom
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
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
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
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
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

	if err := enqueueFilesystemEventJob(context, root, JobPriorityHigh); err != nil {
		t.Fatalf("enqueueFilesystemEventJob returned error: %v", err)
	}
	if len(repository.jobs) != 2 {
		t.Fatalf("expected two queued jobs, got %d", len(repository.jobs))
	}

	if err := enqueueFilesystemEventJob(context, "", JobPriorityLow); err != nil {
		t.Fatalf("expected empty root enqueue to be ignored, got %v", err)
	}

	startupStepPayload, _ := marshalPayload(StepFilePayload{Path: root})
	if string(repository.steps[1].Payload) == "" || string(startupStepPayload) == "" {
		t.Fatalf("expected persisted step payloads")
	}

	if _, err := buildScanPlan("", JobTypeStartupScan, JobPriorityLow); err != nil {
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
