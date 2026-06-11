package engine

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
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
