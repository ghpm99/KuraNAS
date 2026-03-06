package worker

import (
	"os"
	"path/filepath"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"
)

type fileSnapshot struct {
	ModTimeUnix int64
	Size        int64
	IsDir       bool
}

func startEntryPointWatcher(context *WorkerContext) {
	entryPoint := config.AppConfig.EntryPoint
	if entryPoint == "" {
		return
	}

	go func() {
		lastSnapshot := collectEntryPointSnapshot(entryPoint)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		lastDispatchAt := time.Time{}
		pendingChanges := map[string]struct{}{}
		debounceWindow := 2 * time.Second

		for range ticker.C {
			currentSnapshot := collectEntryPointSnapshot(entryPoint)
			changedPaths := snapshotDiffPaths(lastSnapshot, currentSnapshot)
			for _, changedPath := range changedPaths {
				pendingChanges[changedPath] = struct{}{}
			}
			lastSnapshot = currentSnapshot

			if len(pendingChanges) == 0 {
				continue
			}
			if !lastDispatchAt.IsZero() && time.Since(lastDispatchAt) < debounceWindow {
				continue
			}

			if context != nil && context.JobOrchestrator != nil {
				_ = enqueueFilesystemEventJob(context, entryPoint, JobPriorityNormal)
			} else {
				select {
				case context.Tasks <- utils.Task{Type: utils.ScanFiles, Data: "filesystem watch detected changes"}:
				default:
				}
			}
			lastDispatchAt = time.Now()
			pendingChanges = map[string]struct{}{}
		}
	}()
}

func collectEntryPointSnapshot(entryPoint string) map[string]fileSnapshot {
	snapshot := map[string]fileSnapshot{}

	_ = filepath.WalkDir(entryPoint, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}

		snapshot[path] = fileSnapshot{
			ModTimeUnix: info.ModTime().UnixNano(),
			Size:        info.Size(),
			IsDir:       d.IsDir(),
		}

		return nil
	})

	return snapshot
}

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

func snapshotDiffPaths(previous map[string]fileSnapshot, current map[string]fileSnapshot) []string {
	changed := map[string]struct{}{}

	for path, previousSnapshot := range previous {
		currentSnapshot, exists := current[path]
		if !exists || currentSnapshot != previousSnapshot {
			changed[path] = struct{}{}
		}
	}

	for path, currentSnapshot := range current {
		previousSnapshot, exists := previous[path]
		if !exists || previousSnapshot != currentSnapshot {
			changed[path] = struct{}{}
		}
	}

	result := make([]string, 0, len(changed))
	for path := range changed {
		result = append(result, path)
	}

	return result
}
