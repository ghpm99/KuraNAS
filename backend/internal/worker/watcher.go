package worker

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"
)

type fileSnapshot struct {
	ModTimeUnix int64
	Size        int64
	IsDir       bool
}

const watcherMaxIndividualJobs = 50

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
		debounceWindow := 2 * time.Second

		for range ticker.C {
			currentSnapshot := collectEntryPointSnapshot(entryPoint)
			changed := snapshotDiffPaths(lastSnapshot, currentSnapshot)
			lastSnapshot = currentSnapshot

			if len(changed) == 0 {
				continue
			}
			if !lastDispatchAt.IsZero() && time.Since(lastDispatchAt) < debounceWindow {
				continue
			}

			if context != nil && context.JobOrchestrator != nil {
				dispatchWatcherChanges(context, entryPoint, changed, currentSnapshot)
			} else {
				select {
				case context.Tasks <- utils.Task{Type: utils.ScanFiles, Data: "filesystem watch detected changes"}:
				default:
				}
			}
			lastDispatchAt = time.Now()
		}
	}()
}

func dispatchWatcherChanges(context *WorkerContext, entryPoint string, changed []string, currentSnapshot map[string]fileSnapshot) {
	// If too many changes, fall back to a full scan job
	if len(changed) > watcherMaxIndividualJobs {
		if err := enqueueFilesystemEventJob(context, entryPoint, JobPriorityNormal); err != nil {
			log.Printf("[watcher] failed to enqueue full fs_event job: %v\n", err)
		}
		return
	}

	for _, path := range changed {
		snap, existsInCurrent := currentSnapshot[path]

		if !existsInCurrent {
			// File was deleted — enqueue a targeted mark_deleted job
			payload, err := marshalPayload(StepFilePayload{Path: path})
			if err != nil {
				log.Printf("[watcher] failed to marshal mark_deleted payload for %q: %v\n", path, err)
				continue
			}
			if _, err := context.JobOrchestrator.CreateJob(PlannedJob{
				Type:     JobTypeFSEvent,
				Priority: JobPriorityNormal,
				Scope:    JobScope{Path: path},
				Steps: []PlannedStep{
					{
						Key:         "mark_deleted",
						Type:        StepTypeMarkDeleted,
						MaxAttempts: 1,
						Payload:     payload,
					},
				},
			}); err != nil {
				log.Printf("[watcher] failed to create mark_deleted job for %q: %v\n", path, err)
			}
			continue
		}

		// Skip directories
		if snap.IsDir {
			continue
		}

		// New or modified file — create a file processing job
		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}

		fileDto := files.FileDto{
			Path:       path,
			ParentPath: filepath.Dir(path),
		}
		if parseErr := fileDto.ParseFileInfoToFileDto(info); parseErr != nil {
			continue
		}

		plan, planErr := buildFileProcessingPlan(fileDto, JobTypeFSEvent, JobPriorityNormal)
		if planErr != nil {
			log.Printf("[watcher] failed to build plan for %q: %v\n", path, planErr)
			continue
		}

		if _, err := context.JobOrchestrator.CreateJob(plan); err != nil {
			log.Printf("[watcher] failed to create job for %q: %v\n", path, err)
		}
	}
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
