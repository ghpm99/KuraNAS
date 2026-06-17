package engine

import (
	"log"
	"os"
	"path/filepath"

	"nas-go/api/internal/api/v1/captures"
	"nas-go/api/internal/api/v1/trash"

	"github.com/fsnotify/fsnotify"
)

// recursiveWatcher turns fsnotify's per-directory watches into a recursive
// watch over a root tree. inotify (Linux) and fsnotify's portable use of
// ReadDirectoryChangesW (Windows) are both non-recursive, so it keeps one
// watch per directory: the initial tree is walked once at start, and
// directories created later are watched as their Create events arrive.
//
// Consumers read changed paths from Events; whether a path was created,
// modified or removed is resolved by the consumer against the filesystem at
// dispatch time (the state may have flipped since the event fired). Any
// watcher error (inotify queue overflow, RDCW buffer overrun, ...) means
// events may have been lost and is reported through onError so the consumer
// can fall back to a full reconciliation scan.
type recursiveWatcher struct {
	root    string
	watcher *fsnotify.Watcher
	events  chan string
	onError func(error)
	done    chan struct{}
}

func newRecursiveWatcher(root string, onError func(error)) (*recursiveWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	recursive := &recursiveWatcher{
		root:    root,
		watcher: watcher,
		events:  make(chan string, 1024),
		onError: onError,
		done:    make(chan struct{}),
	}

	if err := recursive.watchTree(root, false); err != nil {
		watcher.Close()
		return nil, err
	}

	go recursive.run()
	return recursive, nil
}

// Events emits the paths that changed on disk, one per event.
func (rw *recursiveWatcher) Events() <-chan string {
	return rw.events
}

func (rw *recursiveWatcher) Close() error {
	close(rw.done)
	return rw.watcher.Close()
}

// watchTree adds a watch for dir and every directory below it. With emitPaths
// every path found is also emitted as an event — used for directories created
// after the watcher started, whose contents may have appeared before the
// watch was in place and therefore produced no events of their own.
func (rw *recursiveWatcher) watchTree(dir string, emitPaths bool) error {
	return filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			// Never watch the trash dir: restores/purges inside it are not
			// library changes and must not feed the indexing pipeline.
			if trash.IsInsideTrash(rw.root, path) {
				return filepath.SkipDir
			}
			// Never watch the capture upload staging dir: an in-progress chunked
			// upload rewrites payload.bin every chunk, and indexing the growing
			// partial file repeatedly leaks memory and floods the job queue.
			if captures.IsInsideUploadStaging(rw.root, path) {
				return filepath.SkipDir
			}
			if watchErr := rw.watcher.Add(path); watchErr != nil {
				log.Printf("[watcher] failed to watch %q: %v\n", path, watchErr)
			}
		}
		if emitPaths && path != dir {
			rw.emit(path)
		}
		return nil
	})
}

func (rw *recursiveWatcher) emit(path string) {
	select {
	case rw.events <- path:
	case <-rw.done:
	}
}

func (rw *recursiveWatcher) run() {
	for {
		select {
		case <-rw.done:
			return
		case event, ok := <-rw.watcher.Events:
			if !ok {
				return
			}
			rw.handleEvent(event)
		case err, ok := <-rw.watcher.Errors:
			if !ok {
				return
			}
			rw.handleError(err)
		}
	}
}

func (rw *recursiveWatcher) handleEvent(event fsnotify.Event) {
	// Chmod-only events carry no content change.
	if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
		return
	}

	// A delete-to-trash shows up as a Create inside the trash dir; surfacing
	// it would make the pipeline re-index what the user just deleted.
	if trash.IsInsideTrash(rw.root, event.Name) {
		return
	}

	// In-progress chunked uploads live under capturas/.uploads and rewrite their
	// payload every chunk; never feed those writes to the indexing pipeline.
	if captures.IsInsideUploadStaging(rw.root, event.Name) {
		return
	}

	if event.Op&fsnotify.Create != 0 {
		if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
			// New directory: watch it, and surface anything already inside it
			// (content may have raced ahead of the watch being added).
			if err := rw.watchTree(event.Name, true); err != nil {
				rw.handleError(err)
			}
		}
	}

	rw.emit(event.Name)
}

// handleError reports a watcher failure. An error means events may have been
// dropped, so the consumer must reconcile with a full scan; fsnotify keeps
// the existing watches alive, so event watching itself continues.
func (rw *recursiveWatcher) handleError(err error) {
	if err == nil {
		return
	}
	log.Printf("[watcher] filesystem watcher error: %v\n", err)
	if rw.onError != nil {
		rw.onError(err)
	}
}
