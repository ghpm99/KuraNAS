package engine

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/utils"
)

func TestDispatchWatcherChangesResolvesAgainstDisk(t *testing.T) {
	root := t.TempDir()
	nestedDir := filepath.Join(root, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filePath := filepath.Join(nestedDir, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	// A vanished path becomes mark_deleted, an existing file becomes a
	// processing job, and a directory (FilesService unset here) is skipped.
	deletedPath := filepath.Join(root, "deleted.jpg")
	dispatchWatcherChanges(context, root, []string{deletedPath, nestedDir, filePath})

	if len(repository.jobs) != 2 {
		t.Fatalf("expected mark_deleted and file-processing jobs, got %d", len(repository.jobs))
	}

	overflow := make([]string, watcherMaxIndividualJobs+1)
	for index := range overflow {
		overflow[index] = filepath.Join(root, "overflow", string(rune('a'+(index%26))))
	}
	dispatchWatcherChanges(context, root, overflow)
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

	dispatchWatcherChanges(context, root, []string{root, nestedDir, deepDir, movedDir})

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

func TestEventBatcherDebounceDefersWithoutDropping(t *testing.T) {
	window := 2 * time.Second
	batcher := newEventBatcher(window)
	t0 := time.Now()

	// First batch dispatches on the first flush.
	batcher.add("/a")
	first := batcher.flush(t0)
	if len(first) != 1 || first[0] != "/a" {
		t.Fatalf("expected first batch [/a], got %v", first)
	}

	// Changes landing inside the debounce window are deferred, never dropped.
	batcher.add("/b")
	if got := batcher.flush(t0.Add(time.Second)); got != nil {
		t.Fatalf("expected debounced flush to dispatch nothing, got %v", got)
	}
	batcher.add("/c")
	if got := batcher.flush(t0.Add(1500 * time.Millisecond)); got != nil {
		t.Fatalf("expected debounced flush to dispatch nothing, got %v", got)
	}

	// First allowed flush delivers everything accumulated.
	flushed := batcher.flush(t0.Add(3500 * time.Millisecond))
	flushedSet := map[string]bool{}
	for _, path := range flushed {
		flushedSet[path] = true
	}
	if len(flushed) != 2 || !flushedSet["/b"] || !flushedSet["/c"] {
		t.Fatalf("expected deferred batch [/b /c], got %v", flushed)
	}

	// Nothing pending afterwards.
	if got := batcher.flush(t0.Add(10 * time.Second)); got != nil {
		t.Fatalf("expected quiet flush to dispatch nothing, got %v", got)
	}
}

func TestWatcherVanishedPathYieldsMarkDeletedNotPersist(t *testing.T) {
	root := t.TempDir()
	ghostPath := filepath.Join(root, "ghost.txt")

	// File appeared and vanished before the debounced dispatch ran. Dispatch
	// resolves against the CURRENT disk state: the vanished file must become
	// a mark_deleted job, never an orphan persist job.
	batcher := newEventBatcher(time.Second)
	batcher.add(ghostPath)
	flushed := batcher.flush(time.Now())
	if len(flushed) != 1 || flushed[0] != ghostPath {
		t.Fatalf("expected ghost path flushed, got %v", flushed)
	}

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	dispatchWatcherChanges(context, root, flushed)

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

func TestStartEntryPointWatcherWithoutEntryPointIsNoop(t *testing.T) {
	previous := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = previous })

	config.AppConfig.EntryPoint = ""
	startEntryPointWatcher(&WorkerContext{})
}

func TestWatcherDispatchLoopDeliversNativeEventsAsJobs(t *testing.T) {
	root := t.TempDir()

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	watcher, err := newRecursiveWatcher(root, nil)
	if err != nil {
		t.Fatalf("newRecursiveWatcher: %v", err)
	}
	defer watcher.Close()

	go watcherDispatchLoop(context, root, watcher)

	filePath := filepath.Join(root, "incoming.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Event → batcher → debounced flush → fs_event job, end to end.
	deadline := time.After(5 * time.Second)
	for {
		repository.mu.Lock()
		jobsCount := len(repository.jobs)
		repository.mu.Unlock()
		if jobsCount >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("expected a job dispatched from native watcher event")
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func TestReconcileIntervalIsConfigurableWithSaneDefault(t *testing.T) {
	previous := config.AppConfig.WatcherReconcileHours
	t.Cleanup(func() { config.AppConfig.WatcherReconcileHours = previous })

	config.AppConfig.WatcherReconcileHours = 6
	if got := reconcileInterval(); got != 6*time.Hour {
		t.Fatalf("expected 6h interval, got %v", got)
	}

	config.AppConfig.WatcherReconcileHours = 0
	if got := reconcileInterval(); got != 24*time.Hour {
		t.Fatalf("expected 24h default, got %v", got)
	}
}

func TestWatcherErrorFallbackEnqueuesFullReconciliation(t *testing.T) {
	root := t.TempDir()

	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)
	context := &WorkerContext{JobOrchestrator: orchestrator}

	// Same fallback startEntryPointWatcher wires into the recursive watcher.
	onWatcherError := func(error) {
		if err := enqueueFilesystemEventJob(context, root, job.JobPriorityNormal); err != nil {
			t.Fatalf("enqueueFilesystemEventJob: %v", err)
		}
	}

	watcher, err := newRecursiveWatcher(root, onWatcherError)
	if err != nil {
		t.Fatalf("newRecursiveWatcher: %v", err)
	}
	defer watcher.Close()

	watcher.handleError(os.ErrInvalid)

	if len(repository.jobs) != 1 {
		t.Fatalf("expected a full reconciliation job after watcher error, got %d", len(repository.jobs))
	}
	hasScanStep := false
	for _, step := range repository.steps {
		if step.Type == string(job.StepTypeScanFilesystem) {
			hasScanStep = true
		}
	}
	if !hasScanStep {
		t.Fatalf("expected reconciliation job to carry a scan_filesystem step")
	}
}
