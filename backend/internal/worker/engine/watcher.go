package engine

import (
	"log"
	"nas-go/api/internal/worker/job"
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

// watcherState is the per-tick state machine of the entry-point watcher. The
// debounce may only DEFER a dispatch, never drop it: changes seen during the
// debounce window accumulate in pending and go out on the first allowed tick.
// (The old loop advanced the snapshot before the debounce check, so changes
// landing inside the window were silently lost until a full rescan.)
type watcherState struct {
	lastSnapshot   map[string]fileSnapshot
	pending        map[string]struct{}
	lastDispatchAt time.Time
	debounceWindow time.Duration
}

func newWatcherState(initialSnapshot map[string]fileSnapshot, debounceWindow time.Duration) *watcherState {
	return &watcherState{
		lastSnapshot:   initialSnapshot,
		pending:        map[string]struct{}{},
		debounceWindow: debounceWindow,
	}
}

// processTick merges this tick's changes into the pending set and returns the
// paths to dispatch now — nil while debounced or when nothing is pending. Each
// returned path must be resolved against the CURRENT snapshot by the caller
// (created × deleted may have flipped since the change was first seen).
func (s *watcherState) processTick(currentSnapshot map[string]fileSnapshot, now time.Time) []string {
	for _, path := range snapshotDiffPaths(s.lastSnapshot, currentSnapshot) {
		s.pending[path] = struct{}{}
	}
	s.lastSnapshot = currentSnapshot

	if len(s.pending) == 0 {
		return nil
	}
	if !s.lastDispatchAt.IsZero() && now.Sub(s.lastDispatchAt) < s.debounceWindow {
		return nil
	}

	toDispatch := make([]string, 0, len(s.pending))
	for path := range s.pending {
		toDispatch = append(toDispatch, path)
	}
	s.pending = map[string]struct{}{}
	s.lastDispatchAt = now
	return toDispatch
}

func startEntryPointWatcher(context *WorkerContext) {
	entryPoint := config.AppConfig.EntryPoint
	if entryPoint == "" {
		return
	}

	go func() {
		state := newWatcherState(collectEntryPointSnapshot(entryPoint), 2*time.Second)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			currentSnapshot := collectEntryPointSnapshot(entryPoint)
			changed := state.processTick(currentSnapshot, time.Now())

			if len(changed) == 0 {
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
		}
	}()
}

func dispatchWatcherChanges(context *WorkerContext, entryPoint string, changed []string, currentSnapshot map[string]fileSnapshot) {
	// If too many changes, fall back to a full scan job
	if len(changed) > watcherMaxIndividualJobs {
		if err := enqueueFilesystemEventJob(context, entryPoint, job.JobPriorityNormal); err != nil {
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
				Type:     job.JobTypeFSEvent,
				Priority: job.JobPriorityNormal,
				Scope:    job.JobScope{Path: path},
				Steps: []PlannedStep{
					{
						Key:         "mark_deleted",
						Type:        job.StepTypeMarkDeleted,
						MaxAttempts: 1,
						Payload:     payload,
					},
				},
			}); err != nil {
				log.Printf("[watcher] failed to create mark_deleted job for %q: %v\n", path, err)
			}
			continue
		}

		// New or renamed directory — persist its row directly so it becomes
		// navigable in the tree; directories have no processing plan. The
		// entry point itself stays implicit (the tree lists its children).
		if snap.IsDir {
			if path == entryPoint || context.FilesService == nil {
				continue
			}
			info, statErr := os.Stat(path)
			if statErr != nil {
				continue
			}
			if err := persistDirectoryRow(context.FilesService, path, info); err != nil {
				log.Printf("[watcher] failed to persist directory row for %q: %v\n", path, err)
			}
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

		plan, planErr := buildFileProcessingPlan(fileDto, job.JobTypeFSEvent, job.JobPriorityNormal)
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
