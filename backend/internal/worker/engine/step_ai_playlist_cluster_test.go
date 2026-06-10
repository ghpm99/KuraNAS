package engine

import (
	"context"
	"errors"
	"nas-go/api/internal/worker/job"
	"testing"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/music"
)

// fakeMusicService satisfies music.ServiceInterface; only RebuildAIClusters is
// exercised by the AI playlist clustering step/job, so the rest stays nil.
type fakeMusicService struct {
	music.ServiceInterface
	calls   int
	rebuild error
}

func (f *fakeMusicService) RebuildAIClusters(_ context.Context) error {
	f.calls++
	return f.rebuild
}

func TestExecuteAIPlaylistClusterStepSkipsWithoutMusicService(t *testing.T) {
	if err := executeAIPlaylistClusterStep(nil, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped for nil context, got %v", err)
	}
	if err := executeAIPlaylistClusterStep(&WorkerContext{}, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped without music service, got %v", err)
	}
}

func TestExecuteAIPlaylistClusterStepRebuilds(t *testing.T) {
	svc := &fakeMusicService{}
	if err := executeAIPlaylistClusterStep(&WorkerContext{MusicService: svc}, jobs.StepModel{}); err != nil {
		t.Fatalf("executeAIPlaylistClusterStep error: %v", err)
	}
	if svc.calls != 1 {
		t.Fatalf("expected RebuildAIClusters called once, got %d", svc.calls)
	}

	failing := &fakeMusicService{rebuild: errors.New("model offline")}
	if err := executeAIPlaylistClusterStep(&WorkerContext{MusicService: failing}, jobs.StepModel{}); err == nil {
		t.Fatalf("expected error to propagate from RebuildAIClusters")
	}
}

func TestEnqueueAIPlaylistClusterJobNoOp(t *testing.T) {
	// No orchestrator wired in.
	if err := enqueueAIPlaylistClusterJob(&WorkerContext{MusicService: &fakeMusicService{}}); err != nil {
		t.Fatalf("expected no-op without orchestrator, got %v", err)
	}
	// Orchestrator present but no music service.
	orchestrator := NewJobOrchestrator(newFakeJobsRepository(), nil)
	if err := enqueueAIPlaylistClusterJob(&WorkerContext{JobOrchestrator: orchestrator}); err != nil {
		t.Fatalf("expected no-op without music service, got %v", err)
	}
}

func TestEnqueueAIPlaylistClusterJobCreatesJob(t *testing.T) {
	repo := newFakeJobsRepository()
	ctx := &WorkerContext{
		JobOrchestrator: NewJobOrchestrator(repo, nil),
		MusicService:    &fakeMusicService{},
	}
	if err := enqueueAIPlaylistClusterJob(ctx); err != nil {
		t.Fatalf("enqueueAIPlaylistClusterJob error: %v", err)
	}
	if len(repo.jobs) != 1 {
		t.Fatalf("expected exactly one job enqueued, got %d", len(repo.jobs))
	}
	for _, j := range repo.jobs {
		if j.Type != string(job.JobTypeAIPlaylistCluster) {
			t.Fatalf("expected job type %q, got %q", job.JobTypeAIPlaylistCluster, j.Type)
		}
	}
}
