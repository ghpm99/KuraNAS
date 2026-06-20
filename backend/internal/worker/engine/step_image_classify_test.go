package engine

import (
	"errors"
	"testing"

	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
)

type fakeClassifyImageRepo struct {
	imagedom.RepositoryInterface
	pages   [][]imagedom.PendingImageClassification
	calls   int
	listErr error
}

func (f *fakeClassifyImageRepo) ListPendingAIClassification(threshold float64, after int, limit int) ([]imagedom.PendingImageClassification, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.calls >= len(f.pages) {
		return nil, nil
	}
	page := f.pages[f.calls]
	f.calls++
	return page, nil
}

func newClassifyContext(repo imagedom.RepositoryInterface, settings AISettingsReader) *WorkerContext {
	return &WorkerContext{
		ImageRepository: repo,
		JobOrchestrator: NewJobOrchestrator(newFakeJobsRepository(), nil),
		AISettings:      settings,
	}
}

func TestExecuteImageClassifyEnumerateStep_NilContext(t *testing.T) {
	if err := executeImageClassifyEnumerateStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestExecuteImageClassifyEnumerateStep_ToggleOff(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{}, stubAISettings{enabled: false})
	if err := executeImageClassifyEnumerateStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when toggle off, got %v", err)
	}
}

func TestExecuteImageClassifyEnumerateStep_ToggleError(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{}, stubAISettings{enabled: true, err: errors.New("boom")})
	if err := executeImageClassifyEnumerateStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when toggle read fails")
	}
}

func TestExecuteImageClassifyEnumerateStep_NoPending(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{}, stubAISettings{enabled: true})
	if err := executeImageClassifyEnumerateStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when nothing pending, got %v", err)
	}
}

func TestExecuteImageClassifyEnumerateStep_ListError(t *testing.T) {
	ctx := newClassifyContext(&fakeClassifyImageRepo{listErr: errors.New("boom")}, stubAISettings{enabled: true})
	if err := executeImageClassifyEnumerateStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when listing fails")
	}
}

func TestExecuteImageClassifyEnumerateStep_EnqueuesJobs(t *testing.T) {
	repo := &fakeClassifyImageRepo{
		pages: [][]imagedom.PendingImageClassification{
			{
				{FileID: 1, Path: "/a.jpg"},
				{FileID: 2, Path: "/b.jpg"},
			},
		},
	}
	ctx := newClassifyContext(repo, stubAISettings{enabled: true})
	if err := executeImageClassifyEnumerateStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildImageClassifyMetadataPlan(t *testing.T) {
	plan, err := buildImageClassifyMetadataPlan(imagedom.PendingImageClassification{FileID: 7, Path: "/c.jpg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Type != job.JobTypeFSEvent || plan.Priority != job.JobPriorityLow {
		t.Fatalf("unexpected plan type/priority: %+v", plan)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Type != job.StepTypeMetadata {
		t.Fatalf("expected single metadata step, got %+v", plan.Steps)
	}
	if plan.Scope.Path != "/c.jpg" || plan.Scope.FileID == nil || *plan.Scope.FileID != 7 {
		t.Fatalf("unexpected scope: %+v", plan.Scope)
	}
}
