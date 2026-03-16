package worker

import (
	"os"
	"path/filepath"
	"testing"
)

func snapshotsChanged(previous map[string]fileSnapshot, current map[string]fileSnapshot) bool {
	if len(previous) != len(current) {
		return true
	}

	for path, previousSnapshot := range previous {
		currentSnapshot, exists := current[path]
		if !exists {
			return true
		}
		if currentSnapshot != previousSnapshot {
			return true
		}
	}

	return false
}

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

func TestSnapshotDiffPaths(t *testing.T) {
	previous := map[string]fileSnapshot{
		"/a": {ModTimeUnix: 1, Size: 10, IsDir: false},
		"/b": {ModTimeUnix: 1, Size: 20, IsDir: false},
	}
	current := map[string]fileSnapshot{
		"/a": {ModTimeUnix: 2, Size: 10, IsDir: false},
		"/c": {ModTimeUnix: 1, Size: 30, IsDir: false},
	}

	diff := snapshotDiffPaths(previous, current)
	if len(diff) != 3 {
		t.Fatalf("expected 3 changed paths, got %d (%v)", len(diff), diff)
	}
}

func TestCollectEntryPointSnapshotAndDispatchWatcherChanges(t *testing.T) {
	root := t.TempDir()
	nestedDir := filepath.Join(root, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filePath := filepath.Join(nestedDir, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	snapshot := collectEntryPointSnapshot(root)
	if len(snapshot) < 3 {
		t.Fatalf("expected snapshot to include root, dir, and file: %+v", snapshot)
	}
	if snapshot[filePath].IsDir {
		t.Fatalf("expected file snapshot for %s", filePath)
	}
	if !snapshot[nestedDir].IsDir {
		t.Fatalf("expected directory snapshot for %s", nestedDir)
	}

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	deletedPath := filepath.Join(root, "deleted.jpg")
	dispatchWatcherChanges(
		context,
		root,
		[]string{deletedPath, nestedDir, filePath},
		snapshot,
	)

	if len(repository.jobs) != 2 {
		t.Fatalf("expected mark_deleted and file-processing jobs, got %d", len(repository.jobs))
	}

	overflow := make([]string, watcherMaxIndividualJobs+1)
	for index := range overflow {
		overflow[index] = filepath.Join(root, "overflow", string(rune('a'+(index%26))))
	}
	dispatchWatcherChanges(context, root, overflow, snapshot)
	if len(repository.jobs) != 3 {
		t.Fatalf("expected fallback full-scan job, got %d jobs", len(repository.jobs))
	}
}
