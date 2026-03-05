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

		for range ticker.C {
			currentSnapshot := collectEntryPointSnapshot(entryPoint)
			if snapshotsChanged(lastSnapshot, currentSnapshot) {
				select {
				case context.Tasks <- utils.Task{Type: utils.ScanFiles, Data: "filesystem watch detected changes"}:
				default:
				}
				lastSnapshot = currentSnapshot
			}
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
