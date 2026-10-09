package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/utils"
)

type fakeClassifyImageRepo struct {
	imagedom.RepositoryInterface
	pendingPages  [][]imagedom.PendingImageClassification
	pageRequests  int
	listErr       error
	stored        map[int]imagedom.ClassificationModel
	storeErr      error
	metadataByID  map[int]imagedom.MetadataModel
	missingByPage [][]imagedom.ImageWithoutMetadata
	missingCalls  int
}

func (f *fakeClassifyImageRepo) ListPendingAIClassification(threshold float64, after int, limit int) ([]imagedom.PendingImageClassification, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.pageRequests >= len(f.pendingPages) {
		return nil, nil
	}
	page := f.pendingPages[f.pageRequests]
	f.pageRequests++
	return page, nil
}

func (f *fakeClassifyImageRepo) GetImageMetadataByID(id int) (imagedom.MetadataModel, error) {
	return f.metadataByID[id], nil
}

func (f *fakeClassifyImageRepo) UpdateAIClassification(fileID int, classification imagedom.ClassificationModel) error {
	if f.storeErr != nil {
		return f.storeErr
	}
	if f.stored == nil {
		f.stored = map[int]imagedom.ClassificationModel{}
	}
	f.stored[fileID] = classification
	return nil
}

func (f *fakeClassifyImageRepo) ListImagesWithoutMetadata(after int, limit int) ([]imagedom.ImageWithoutMetadata, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.missingCalls >= len(f.missingByPage) {
		return nil, nil
	}
	page := f.missingByPage[f.missingCalls]
	f.missingCalls++
	return page, nil
}

type fakeVisionService struct {
	calls     atomic.Int32
	executeFn func(ctx context.Context, req ai.Request) (ai.Response, error)
}

func (f *fakeVisionService) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	f.calls.Add(1)
	if f.executeFn != nil {
		return f.executeFn(ctx, req)
	}
	return ai.Response{Content: `{"category": "landscape", "confidence": 0.9}`}, nil
}

func newClassifyContext(repo imagedom.RepositoryInterface, vision ai.ServiceInterface, settings AISettingsReader) *WorkerContext {
	return &WorkerContext{
		ImageRepository: repo,
		FilesService:    &workerFilesServiceMock{},
		JobOrchestrator: NewJobOrchestrator(newFakeJobsRepository(), nil),
		AIService:       vision,
		AISettings:      settings,
	}
}

func pendingImages(fileIDs ...int) []imagedom.PendingImageClassification {
	pending := make([]imagedom.PendingImageClassification, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		pending = append(pending, imagedom.PendingImageClassification{FileID: fileID, Path: "/img.jpg", MetadataID: fileID * 10})
	}
	return pending
}

func TestExecuteImageClassifyBatchStep_RequiresDependencies(t *testing.T) {
	if err := executeImageClassifyBatchStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestExecuteImageClassifyBatchStep_ToggleOffNeverCallsAI(t *testing.T) {
	vision := &fakeVisionService{}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{pendingImages(1)}}
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: false})

	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when toggle off, got %v", err)
	}
	if vision.calls.Load() != 0 || len(repo.stored) != 0 {
		t.Fatalf("expected no AI calls and no stored classification, calls=%d", vision.calls.Load())
	}
}

func TestExecuteImageClassifyBatchStep_NothingPendingIsSkipped(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{}, &fakeVisionService{}, stubAISettings{enabled: true})
	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteImageClassifyBatchStep_ListErrorFails(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{listErr: errors.New("boom")}, &fakeVisionService{}, stubAISettings{enabled: true})
	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when listing fails")
	}
}

func TestExecuteImageClassifyBatchStep_StoresAIClassification(t *testing.T) {
	vision := &fakeVisionService{}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{pendingImages(1, 2)}}
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: true})

	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vision.calls.Load() != 2 || len(repo.stored) != 2 {
		t.Fatalf("expected 2 AI calls and 2 stored, calls=%d stored=%d", vision.calls.Load(), len(repo.stored))
	}
	if repo.stored[1].Category != imagedom.ClassificationCategoryLandscape {
		t.Fatalf("unexpected stored classification: %+v", repo.stored[1])
	}
}

func TestExecuteImageClassifyBatchStep_PagesThroughLargeBacklogInOneStep(t *testing.T) {
	fullPage := make([]imagedom.PendingImageClassification, 0, imageClassifyPageSize)
	for fileID := 1; fileID <= imageClassifyPageSize; fileID++ {
		fullPage = append(fullPage, imagedom.PendingImageClassification{FileID: fileID, Path: "/img.jpg", MetadataID: fileID})
	}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{fullPage, pendingImages(1001)}}
	vision := &fakeVisionService{}
	jobsRepository := newFakeJobsRepository()
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: true})
	ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)

	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(vision.calls.Load()) != imageClassifyPageSize+1 {
		t.Fatalf("expected %d AI calls, got %d", imageClassifyPageSize+1, vision.calls.Load())
	}
	if len(jobsRepository.jobs) != 0 {
		t.Fatalf("backlog paging must not create jobs, got %d", len(jobsRepository.jobs))
	}
}

func TestExecuteImageClassifyBatchStep_CallTimeoutIsRecordedWithoutStoring(t *testing.T) {
	previousTimeout := imageClassifyCallTimeout
	imageClassifyCallTimeout = 20 * time.Millisecond
	t.Cleanup(func() { imageClassifyCallTimeout = previousTimeout })

	vision := &fakeVisionService{executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
		<-ctx.Done()
		return ai.Response{}, ctx.Err()
	}}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{pendingImages(1, 2)}}
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: true})

	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("timeouts below the abort threshold must not fail the step, got %v", err)
	}
	if vision.calls.Load() != 2 {
		t.Fatalf("expected both images attempted, got %d calls", vision.calls.Load())
	}
	if len(repo.stored) != 0 {
		t.Fatalf("timed out images must not be stored, got %v", repo.stored)
	}
}

func TestExecuteImageClassifyBatchStep_AbortsAfterConsecutiveFailures(t *testing.T) {
	vision := &fakeVisionService{executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
		return ai.Response{}, errors.New("provider down")
	}}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{pendingImages(1, 2, 3, 4, 5, 6, 7, 8)}}
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: true})

	err := executeImageClassifyBatchStep(ctx, jobs.StepModel{})
	if !errors.Is(err, errImageClassifyAborted) {
		t.Fatalf("expected abort error, got %v", err)
	}
	if int(vision.calls.Load()) != imageClassifyMaxConsecutiveFailures {
		t.Fatalf("expected %d attempts before aborting, got %d", imageClassifyMaxConsecutiveFailures, vision.calls.Load())
	}
}

func TestExecuteImageClassifyBatchStep_SuccessResetsFailureStreak(t *testing.T) {
	var attempt atomic.Int32
	vision := &fakeVisionService{executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
		if attempt.Add(1)%2 == 1 {
			return ai.Response{}, errors.New("flaky")
		}
		return ai.Response{Content: `{"category": "art", "confidence": 0.8}`}, nil
	}}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{pendingImages(1, 2, 3, 4, 5, 6, 7, 8)}}
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: true})

	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("alternating failures must not abort, got %v", err)
	}
	if len(repo.stored) != 4 {
		t.Fatalf("expected 4 stored classifications, got %d", len(repo.stored))
	}
}

func TestExecuteImageClassifyBatchStep_MissingFileIsSkippedNotFailed(t *testing.T) {
	vision := &fakeVisionService{}
	repo := &fakeClassifyImageRepo{pendingPages: [][]imagedom.PendingImageClassification{pendingImages(1)}}
	ctx := newClassifyContext(repo, vision, stubAISettings{enabled: true})
	ctx.FilesService = &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		return files.FileDto{}, sql.ErrNoRows
	}}

	if err := executeImageClassifyBatchStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
	if vision.calls.Load() != 0 {
		t.Fatal("AI must not run for a file that no longer exists")
	}
}

func TestBuildImageAIClassifyPlan(t *testing.T) {
	plan := buildImageAIClassifyPlan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("plan must be valid: %v", err)
	}
	if plan.Type != job.JobTypeImageClassifyBackfill || plan.Priority != job.JobPriorityLow {
		t.Fatalf("unexpected plan type/priority: %+v", plan)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Type != job.StepTypeImageClassifyBatch {
		t.Fatalf("expected single batch step, got %+v", plan.Steps)
	}
	if plan.Scope.Path != imagedom.AIClassificationJobScopePath {
		t.Fatalf("expected shared scope path for dedupe, got %q", plan.Scope.Path)
	}
}

func TestEnqueueImageAIClassifyIfNeeded(t *testing.T) {
	lowConfidence := imagedom.ClassificationModel{Confidence: 0.35}
	highConfidence := imagedom.ClassificationModel{Confidence: 0.9}

	t.Run("low confidence with AI enabled enqueues one job", func(t *testing.T) {
		jobsRepository := newFakeJobsRepository()
		ctx := newClassifyContext(&fakeClassifyImageRepo{}, &fakeVisionService{}, stubAISettings{enabled: true})
		ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)
		enqueueImageAIClassifyIfNeeded(ctx, lowConfidence, "/a.jpg")
		enqueueImageAIClassifyIfNeeded(ctx, lowConfidence, "/b.jpg")
		if len(jobsRepository.jobs) != 1 {
			t.Fatalf("expected a single deduplicated job, got %d", len(jobsRepository.jobs))
		}
	})

	t.Run("high confidence does not enqueue", func(t *testing.T) {
		jobsRepository := newFakeJobsRepository()
		ctx := newClassifyContext(&fakeClassifyImageRepo{}, &fakeVisionService{}, stubAISettings{enabled: true})
		ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)
		enqueueImageAIClassifyIfNeeded(ctx, highConfidence, "/a.jpg")
		if len(jobsRepository.jobs) != 0 {
			t.Fatalf("expected no job, got %d", len(jobsRepository.jobs))
		}
	})

	t.Run("toggle off never enqueues", func(t *testing.T) {
		jobsRepository := newFakeJobsRepository()
		ctx := newClassifyContext(&fakeClassifyImageRepo{}, &fakeVisionService{}, stubAISettings{enabled: false})
		ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)
		enqueueImageAIClassifyIfNeeded(ctx, lowConfidence, "/a.jpg")
		if len(jobsRepository.jobs) != 0 {
			t.Fatalf("expected no job, got %d", len(jobsRepository.jobs))
		}
	})

	t.Run("nil context is a no-op", func(t *testing.T) {
		enqueueImageAIClassifyIfNeeded(nil, lowConfidence, "/a.jpg")
	})
}

func TestExecuteMetadataStep_ImageNeverCallsAIAndSchedulesBacklogJob(t *testing.T) {
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(imagedom.MetadataModel{Path: filePath, Format: "jpg", Width: 10, Height: 10})
		return string(payload), nil
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })

	vision := &fakeVisionService{}
	var upserted imagedom.MetadataModel
	imageRepository := newFakeEngineImageRepository(func(tx *sql.Tx, metadata imagedom.MetadataModel) (imagedom.MetadataModel, error) {
		upserted = metadata
		return metadata, nil
	})
	jobsRepository := newFakeJobsRepository()
	ctx := newClassifyContext(imageRepository, vision, stubAISettings{enabled: true})
	ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)

	payload, _ := marshalPayload(StepFilePayload{File: &files.FileDto{
		ID: 10, Name: "wallpaper.jpg", Path: "/downloads/wallpaper.jpg", Format: ".jpg", Type: files.File,
	}})
	if err := executeMetadataStep(ctx, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vision.calls.Load() != 0 {
		t.Fatalf("metadata step must not call the AI, got %d calls", vision.calls.Load())
	}
	if upserted.Classification.ClassifiedByAI {
		t.Fatal("metadata step must persist only the heuristic classification")
	}
	if len(jobsRepository.jobs) != 1 {
		t.Fatalf("expected one backlog classification job, got %d", len(jobsRepository.jobs))
	}
}
