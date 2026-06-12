package engine

import (
	"log"
	"nas-go/api/internal/worker/job"
	"os"
	"path/filepath"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
)

const watcherMaxIndividualJobs = 50
const watcherDebounceWindow = 2 * time.Second
const watcherFlushPoll = 500 * time.Millisecond

// eventBatcher coalesces the changed paths reported by the native watcher.
// The debounce may only DEFER a dispatch, never drop it: changes seen during
// the debounce window accumulate in pending and go out on the first allowed
// flush. Each dispatched path must be resolved against the CURRENT disk state
// by the caller (created × deleted may have flipped since the event fired).
type eventBatcher struct {
	pending        map[string]struct{}
	lastDispatchAt time.Time
	debounceWindow time.Duration
}

func newEventBatcher(debounceWindow time.Duration) *eventBatcher {
	return &eventBatcher{
		pending:        map[string]struct{}{},
		debounceWindow: debounceWindow,
	}
}

func (b *eventBatcher) add(path string) {
	b.pending[path] = struct{}{}
}

// flush returns the paths to dispatch now — nil while debounced or when
// nothing is pending.
func (b *eventBatcher) flush(now time.Time) []string {
	if len(b.pending) == 0 {
		return nil
	}
	if !b.lastDispatchAt.IsZero() && now.Sub(b.lastDispatchAt) < b.debounceWindow {
		return nil
	}

	toDispatch := make([]string, 0, len(b.pending))
	for path := range b.pending {
		toDispatch = append(toDispatch, path)
	}
	b.pending = map[string]struct{}{}
	b.lastDispatchAt = now
	return toDispatch
}

// startRootWatchers watches every enabled storage root with OS-native
// filesystem events (fsnotify): near-zero CPU/IO at rest, so a mechanical disk
// can spin down. Full-tree scans remain only as reconciliation — at boot (the
// existing startup_scan) and at a low configurable frequency — plus the
// automatic fallback when a native watcher reports an error/overflow.
func startRootWatchers(context *WorkerContext) {
	enabledRoots := roots.Enabled()
	if len(enabledRoots) == 0 {
		return
	}

	for _, root := range enabledRoots {
		rootPath := root.Path
		onWatcherError := func(error) {
			// Events may have been lost: reconcile this root's whole tree.
			if err := enqueueFilesystemEventJob(context, rootPath, job.JobPriorityNormal); err != nil {
				log.Printf("[watcher] failed to enqueue overflow reconciliation for %q: %v\n", rootPath, err)
			}
		}

		watcher, err := newRecursiveWatcher(rootPath, onWatcherError)
		if err != nil {
			log.Printf("[watcher] native watcher unavailable for %q (%v); relying on reconciliation scans only\n", rootPath, err)
			continue
		}
		go watcherDispatchLoop(context, rootPath, watcher)
	}

	go reconciliationLoop(context)
}

func watcherDispatchLoop(context *WorkerContext, rootPath string, watcher *recursiveWatcher) {
	batcher := newEventBatcher(watcherDebounceWindow)
	ticker := time.NewTicker(watcherFlushPoll)
	defer ticker.Stop()

	for {
		select {
		case path, ok := <-watcher.Events():
			if !ok {
				return
			}
			batcher.add(path)
		case <-ticker.C:
			changed := batcher.flush(time.Now())
			if len(changed) == 0 {
				continue
			}

			if context != nil && context.JobOrchestrator != nil {
				dispatchWatcherChanges(context, rootPath, changed)
			}
		}
	}
}

func reconcileInterval() time.Duration {
	hours := config.AppConfig.WatcherReconcileHours
	if hours <= 0 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}

// reconciliationLoop enqueues a low-priority full-tree fs_event per enabled
// root at the WATCHER_RECONCILE_HOURS interval (default 24h) to capture
// anything the native watchers missed. Boot reconciliation is the existing
// startup_scan. The roots are re-read every tick, so roots registered after
// boot are reconciled too.
func reconciliationLoop(context *WorkerContext) {
	ticker := time.NewTicker(reconcileInterval())
	defer ticker.Stop()

	for range ticker.C {
		for _, root := range roots.Enabled() {
			if err := enqueueFilesystemEventJob(context, root.Path, job.JobPriorityLow); err != nil {
				log.Printf("[watcher] failed to enqueue periodic reconciliation for %q: %v\n", root.Path, err)
			}
		}
	}
}

func dispatchWatcherChanges(context *WorkerContext, rootPath string, changed []string) {
	// If too many changes, fall back to a full scan job
	if len(changed) > watcherMaxIndividualJobs {
		if err := enqueueFilesystemEventJob(context, rootPath, job.JobPriorityNormal); err != nil {
			log.Printf("[watcher] failed to enqueue full fs_event job: %v\n", err)
		}
		return
	}

	for _, path := range changed {
		info, statErr := os.Stat(path)

		if statErr != nil {
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
		// navigable in the tree; directories have no processing plan. Storage
		// roots get a row too: they are the top-level nodes of the tree.
		if info.IsDir() {
			if context.FilesService == nil {
				continue
			}
			if err := persistDirectoryRow(context.FilesService, path, info); err != nil {
				log.Printf("[watcher] failed to persist directory row for %q: %v\n", path, err)
			}
			continue
		}

		// New or modified file — create a file processing job
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
