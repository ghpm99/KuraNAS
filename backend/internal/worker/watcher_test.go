package worker

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/worker/domain"
)

func TestSnapshotsChanged(t *testing.T) {
	base := map[string]fileSnapshot{
		"/a": {ModTimeUnix: 1, Size: 10, IsDir: false},
	}

	if snapshotsChanged(base, map[string]fileSnapshot{"/a": {ModTimeUnix: 1, Size: 10, IsDir: false}}) {
		t.Fatalf("expected equal snapshots to be unchanged")
	}

	if !snapshotsChanged(base, map[string]fileSnapshot{"/a": {ModTimeUnix: 2, Size: 10, IsDir: false}}) {
		t.Fatalf("expected change when metadata differs")
	}

	if !snapshotsChanged(base, map[string]fileSnapshot{"/b": {ModTimeUnix: 1, Size: 10, IsDir: false}}) {
		t.Fatalf("expected change when path differs")
	}

	if !snapshotsChanged(base, map[string]fileSnapshot{"/a": {ModTimeUnix: 1, Size: 10, IsDir: false}, "/b": {ModTimeUnix: 1, Size: 1, IsDir: false}}) {
		t.Fatalf("expected change when snapshot sizes differ")
	}
}

func TestEnqueueFilesystemChangeJobCreatesFSEventJob(t *testing.T) {
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	ctx := &WorkerContext{
		Orchestrator: orchestrator,
	}

	enqueueFilesystemChangeJob(ctx, domain.NewRootScopePayload("/tmp"))

	if len(repo.jobsByID) != 1 {
		t.Fatalf("expected one fs_event job to be created, got %d", len(repo.jobsByID))
	}
}

func TestEnqueueFilesystemChangeJobDoesNotFallbackToLegacyQueue(t *testing.T) {
	ctx := &WorkerContext{}

	enqueueFilesystemChangeJob(ctx, domain.NewRootScopePayload("/tmp"))
}

func TestCollectChangedScopes_DeduplicatesAndUsesRootScope(t *testing.T) {
	previous := map[string]fileSnapshot{
		"/data/movie.mp4": {ModTimeUnix: 10, Size: 100, IsDir: false},
	}
	current := map[string]fileSnapshot{
		"/data/movie.mp4": {ModTimeUnix: 11, Size: 100, IsDir: false},
		"/data/new.mp4":   {ModTimeUnix: 1, Size: 10, IsDir: false},
	}

	scopes := collectChangedScopes(previous, current, "/data")
	if len(scopes) != 1 {
		t.Fatalf("expected one deduplicated scope, got %d", len(scopes))
	}
	if scopes[0].Type != domain.ScopeTypeRoot {
		t.Fatalf("expected root scope for root-level file changes, got %s", scopes[0].Type)
	}
	if scopes[0].Root == nil || scopes[0].Root.Root != "/data" {
		t.Fatalf("expected root scope /data, got %+v", scopes[0].Root)
	}
}

func TestCollectChangedScopes_DeletedFileUsesParentPath(t *testing.T) {
	previous := map[string]fileSnapshot{
		"/data/folder/old.mp4": {ModTimeUnix: 10, Size: 100, IsDir: false},
	}
	current := map[string]fileSnapshot{}

	scopes := collectChangedScopes(previous, current, "/data")
	if len(scopes) != 1 {
		t.Fatalf("expected one scope for delete event, got %d", len(scopes))
	}
	if scopes[0].Type != domain.ScopeTypePath {
		t.Fatalf("expected path scope for nested delete, got %s", scopes[0].Type)
	}
	if scopes[0].Path == nil || scopes[0].Path.Path != "/data/folder" {
		t.Fatalf("expected parent path scope /data/folder, got %+v", scopes[0].Path)
	}
}

func TestCollectChangedScopes_RenameProducesSingleParentScope(t *testing.T) {
	previous := map[string]fileSnapshot{
		"/data/folder/old.mp4": {ModTimeUnix: 10, Size: 100, IsDir: false},
	}
	current := map[string]fileSnapshot{
		"/data/folder/new.mp4": {ModTimeUnix: 11, Size: 100, IsDir: false},
	}

	scopes := collectChangedScopes(previous, current, "/data")
	if len(scopes) != 1 {
		t.Fatalf("expected one deduplicated parent scope for rename, got %d", len(scopes))
	}
	if scopes[0].Type != domain.ScopeTypePath {
		t.Fatalf("expected path scope for rename in nested folder, got %s", scopes[0].Type)
	}
	if scopes[0].Path == nil || scopes[0].Path.Path != "/data/folder" {
		t.Fatalf("expected parent path scope /data/folder, got %+v", scopes[0].Path)
	}
}

func TestFilesystemEventDebouncer_FlushesAfterDebounce(t *testing.T) {
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	debouncer := newFilesystemEventDebouncer(
		&WorkerContext{Orchestrator: orchestrator},
		"/data",
		20*time.Millisecond,
		10,
	)
	debouncer.AddScopes([]domain.ScopePayload{
		domain.NewPathScopePayload("/data/a"),
		domain.NewPathScopePayload("/data/a"),
		domain.NewPathScopePayload("/data/b"),
	})

	debouncer.TryFlush()
	if len(repo.jobsByID) != 0 {
		t.Fatalf("did not expect flush before debounce window")
	}

	time.Sleep(25 * time.Millisecond)
	debouncer.TryFlush()
	if len(repo.jobsByID) != 2 {
		t.Fatalf("expected two jobs after debounce flush, got %d", len(repo.jobsByID))
	}
}

func TestFilesystemEventDebouncer_BurstsFallbackToRootScope(t *testing.T) {
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	debouncer := newFilesystemEventDebouncer(
		&WorkerContext{Orchestrator: orchestrator},
		"/data",
		1*time.Millisecond,
		2,
	)
	debouncer.AddScopes([]domain.ScopePayload{
		domain.NewPathScopePayload("/data/a"),
		domain.NewPathScopePayload("/data/b"),
		domain.NewPathScopePayload("/data/c"),
	})

	time.Sleep(2 * time.Millisecond)
	debouncer.TryFlush()
	if len(repo.jobsByID) != 1 {
		t.Fatalf("expected one root-scoped job in burst mode, got %d", len(repo.jobsByID))
	}

	for _, job := range repo.jobsByID {
		if job.ScopeJSON == "" || job.ScopeJSON != `{"type":"root","root":{"root":"/data"}}` {
			t.Fatalf("expected root scope payload, got %s", job.ScopeJSON)
		}
	}
}
