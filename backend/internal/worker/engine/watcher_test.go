package engine

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/utils"
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

func TestDispatchWatcherChangesPersistsDirectories(t *testing.T) {
	root := t.TempDir()
	nestedDir := filepath.Join(root, "nested")
	deepDir := filepath.Join(nestedDir, "deep")
	movedDir := filepath.Join(root, "moved")
	for _, dir := range []string{deepDir, movedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
	}

	snapshot := collectEntryPointSnapshot(root)

	created := []files.FileDto{}
	updated := []files.FileDto{}
	filesService := &workerFilesServiceMock{
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			// movedDir still has a soft-deleted row from before the move.
			if path == movedDir {
				return files.FileDto{
					ID:        7,
					Name:      name,
					Path:      path,
					Type:      files.Directory,
					DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now()},
				}, nil
			}
			return files.FileDto{}, sql.ErrNoRows
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			created = append(created, fileDto)
			return fileDto, nil
		},
		updateFileFn: func(fileDto files.FileDto) (bool, error) {
			updated = append(updated, fileDto)
			return true, nil
		},
	}

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator, FilesService: filesService}

	dispatchWatcherChanges(
		context,
		root,
		[]string{root, nestedDir, deepDir, movedDir},
		snapshot,
	)

	if len(repository.jobs) != 0 {
		t.Fatalf("directories must not enqueue processing jobs, got %d", len(repository.jobs))
	}

	if len(created) != 2 {
		t.Fatalf("expected 2 directory rows created (nested + deep), got %+v", created)
	}
	for _, fileDto := range created {
		if fileDto.Type != files.Directory {
			t.Fatalf("expected directory type for %q, got %v", fileDto.Path, fileDto.Type)
		}
		if fileDto.Path == root {
			t.Fatalf("entry point itself must not get a row")
		}
	}

	if len(updated) != 1 || updated[0].ID != 7 || updated[0].DeletedAt.HasValue {
		t.Fatalf("expected soft-deleted directory row to be revived, got %+v", updated)
	}
}

func TestWatcherStateDebounceDefersWithoutDropping(t *testing.T) {
	snap := func(paths ...string) map[string]fileSnapshot {
		out := map[string]fileSnapshot{}
		for index, path := range paths {
			out[path] = fileSnapshot{ModTimeUnix: int64(index + 1), Size: 10}
		}
		return out
	}

	window := 2 * time.Second
	state := newWatcherState(snap(), window)
	t0 := time.Now()

	// First batch dispatches immediately.
	first := state.processTick(snap("/a"), t0)
	if len(first) != 1 || first[0] != "/a" {
		t.Fatalf("expected first batch [/a], got %v", first)
	}

	// Second and third batches land inside the debounce window: deferred,
	// never dropped.
	if got := state.processTick(snap("/a", "/b"), t0.Add(time.Second)); got != nil {
		t.Fatalf("expected debounced tick to dispatch nothing, got %v", got)
	}
	if got := state.processTick(snap("/a", "/b", "/c"), t0.Add(1500*time.Millisecond)); got != nil {
		t.Fatalf("expected debounced tick to dispatch nothing, got %v", got)
	}

	// First allowed tick flushes everything accumulated, even with no new
	// change in this tick.
	flushed := state.processTick(snap("/a", "/b", "/c"), t0.Add(3500*time.Millisecond))
	flushedSet := map[string]bool{}
	for _, path := range flushed {
		flushedSet[path] = true
	}
	if len(flushed) != 2 || !flushedSet["/b"] || !flushedSet["/c"] {
		t.Fatalf("expected deferred batch [/b /c], got %v", flushed)
	}

	// Nothing pending afterwards.
	if got := state.processTick(snap("/a", "/b", "/c"), t0.Add(10*time.Second)); got != nil {
		t.Fatalf("expected quiet tick to dispatch nothing, got %v", got)
	}
}

func TestWatcherStateVanishedPathYieldsNoPersistJob(t *testing.T) {
	root := t.TempDir()
	ghostPath := filepath.Join(root, "ghost.txt")

	window := 2 * time.Second
	state := newWatcherState(map[string]fileSnapshot{}, window)
	t0 := time.Now()

	// Burn the first allowed dispatch so the next changes fall in the window.
	if got := state.processTick(map[string]fileSnapshot{"/warm": {Size: 1}}, t0); len(got) != 1 {
		t.Fatalf("expected warm-up dispatch, got %v", got)
	}

	// File appears inside the debounce window…
	if got := state.processTick(map[string]fileSnapshot{"/warm": {Size: 1}, ghostPath: {Size: 5}}, t0.Add(time.Second)); got != nil {
		t.Fatalf("expected debounced tick to dispatch nothing, got %v", got)
	}
	// …and vanishes before the dispatch is allowed.
	currentSnapshot := map[string]fileSnapshot{"/warm": {Size: 1}}
	flushed := state.processTick(currentSnapshot, t0.Add(3*time.Second))
	if len(flushed) != 1 || flushed[0] != ghostPath {
		t.Fatalf("expected ghost path flushed, got %v", flushed)
	}

	// Dispatch resolves against the CURRENT snapshot: the vanished file must
	// become a mark_deleted job, never an orphan persist job.
	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	dispatchWatcherChanges(context, root, flushed, currentSnapshot)

	if len(repository.jobs) != 1 {
		t.Fatalf("expected exactly one job for the vanished file, got %d", len(repository.jobs))
	}
	for _, step := range repository.steps {
		if step.Type == string(job.StepTypePersist) {
			t.Fatalf("vanished file must not produce a persist job, got step %+v", step)
		}
	}
	foundMarkDeleted := false
	for _, step := range repository.steps {
		if step.Type == string(job.StepTypeMarkDeleted) {
			foundMarkDeleted = true
		}
	}
	if !foundMarkDeleted {
		t.Fatalf("expected a mark_deleted step for the vanished file")
	}
}
