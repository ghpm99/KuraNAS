package engine

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/utils"
)

func TestExecuteImageMetadataReconcileStep_RequiresDependencies(t *testing.T) {
	if err := executeImageMetadataReconcileStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestExecuteImageMetadataReconcileStep_NothingMissingIsSkipped(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{}, nil, nil)
	if err := executeImageMetadataReconcileStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteImageMetadataReconcileStep_ListErrorFails(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{listErr: errors.New("boom")}, nil, nil)
	if err := executeImageMetadataReconcileStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when listing fails")
	}
}

func TestExecuteImageMetadataReconcileStep_RunsMetadataForEachMissingImageAcrossPages(t *testing.T) {
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(imagedom.MetadataModel{Path: filePath, Format: "jpg", Width: 10, Height: 10})
		return string(payload), nil
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })

	fullPage := make([]imagedom.ImageWithoutMetadata, 0, imageMetadataReconcilePageSize)
	for fileID := 1; fileID <= imageMetadataReconcilePageSize; fileID++ {
		fullPage = append(fullPage, imagedom.ImageWithoutMetadata{FileID: fileID, Path: "/p.jpg"})
	}
	upserts := 0
	imageRepo := newFakeEngineImageRepository(func(tx *sql.Tx, m imagedom.MetadataModel) (imagedom.MetadataModel, error) {
		upserts++
		return m, nil
	})
	for range imageMetadataReconcilePageSize {
		imageRepo.mock.ExpectBegin()
		imageRepo.mock.ExpectCommit()
	}
	reconcileRepo := &reconcileImageRepository{
		fakeEngineImageRepository: imageRepo,
		missingByPage:             [][]imagedom.ImageWithoutMetadata{fullPage, {{FileID: 9999, Path: "/last.jpg"}}},
	}

	jobsRepository := newFakeJobsRepository()
	ctx := newClassifyContext(reconcileRepo, nil, stubAISettings{enabled: false})
	ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)
	ctx.FilesService = &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		return files.FileDto{ID: id, Name: "p.jpg", Path: "/p.jpg", Format: ".jpg"}, nil
	}}

	if err := executeImageMetadataReconcileStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if upserts != imageMetadataReconcilePageSize+1 {
		t.Fatalf("expected %d metadata upserts, got %d", imageMetadataReconcilePageSize+1, upserts)
	}
	if len(jobsRepository.jobs) != 0 {
		t.Fatalf("reconcile must not create per-image jobs, got %d", len(jobsRepository.jobs))
	}
}

func TestExecuteImageMetadataReconcileStep_FailuresAreRecordedWithoutFailingTheStep(t *testing.T) {
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{}", nil
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })

	repo := &fakeClassifyImageRepo{missingByPage: [][]imagedom.ImageWithoutMetadata{{{FileID: 1, Path: "/a.jpg"}, {FileID: 2, Path: "/b.jpg"}}}}
	ctx := newClassifyContext(repo, nil, nil)
	ctx.FilesService = &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		return files.FileDto{}, errors.New("db down")
	}}

	if err := executeImageMetadataReconcileStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("per-file failures must not fail the step, got %v", err)
	}
}

type reconcileImageRepository struct {
	*fakeEngineImageRepository
	missingByPage [][]imagedom.ImageWithoutMetadata
	missingCalls  int
}

func (r *reconcileImageRepository) ListImagesWithoutMetadata(after int, limit int) ([]imagedom.ImageWithoutMetadata, error) {
	if r.missingCalls >= len(r.missingByPage) {
		return nil, nil
	}
	page := r.missingByPage[r.missingCalls]
	r.missingCalls++
	return page, nil
}

func TestBuildImageMetadataReconcilePlan(t *testing.T) {
	plan := buildImageMetadataReconcilePlan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("plan must be valid: %v", err)
	}
	if plan.Type != job.JobTypeImageMetadataReconcile || plan.Priority != job.JobPriorityLow {
		t.Fatalf("unexpected plan type/priority: %+v", plan)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Type != job.StepTypeImageMetadataReconcile {
		t.Fatalf("expected single reconcile step, got %+v", plan.Steps)
	}
}

func TestEnqueueImageMetadataReconcileJob(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	ctx := newClassifyContext(&fakeClassifyImageRepo{}, nil, nil)
	ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)

	if err := enqueueImageMetadataReconcileJob(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := enqueueImageMetadataReconcileJob(ctx); err != nil {
		t.Fatalf("unexpected error on second enqueue: %v", err)
	}
	if len(jobsRepository.jobs) != 1 {
		t.Fatalf("expected one deduplicated reconcile job, got %d", len(jobsRepository.jobs))
	}

	if err := enqueueImageMetadataReconcileJob(nil); err != nil {
		t.Fatalf("nil context must be a no-op, got %v", err)
	}
}
