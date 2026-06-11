package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// collectEventsUntil drains watcher events into a set until want paths were
// seen or the timeout expires.
func collectEventsUntil(t *testing.T, watcher *recursiveWatcher, want map[string]bool, timeout time.Duration) map[string]bool {
	t.Helper()
	seen := map[string]bool{}
	deadline := time.After(timeout)
	for {
		missing := false
		for path := range want {
			if !seen[path] {
				missing = true
				break
			}
		}
		if !missing {
			return seen
		}

		select {
		case path := <-watcher.Events():
			seen[path] = true
		case <-deadline:
			return seen
		}
	}
}

func TestRecursiveWatcherDetectsCreateWriteAndRemove(t *testing.T) {
	root := t.TempDir()
	watcher, err := newRecursiveWatcher(root, nil)
	if err != nil {
		t.Fatalf("newRecursiveWatcher: %v", err)
	}
	defer watcher.Close()

	filePath := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	seen := collectEventsUntil(t, watcher, map[string]bool{filePath: true}, 2*time.Second)
	if !seen[filePath] {
		t.Fatalf("create event for %q not received, saw %v", filePath, seen)
	}

	if err := os.WriteFile(filePath, []byte("image v2 with more bytes"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	seen = collectEventsUntil(t, watcher, map[string]bool{filePath: true}, 2*time.Second)
	if !seen[filePath] {
		t.Fatalf("write event for %q not received, saw %v", filePath, seen)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	seen = collectEventsUntil(t, watcher, map[string]bool{filePath: true}, 2*time.Second)
	if !seen[filePath] {
		t.Fatalf("remove event for %q not received, saw %v", filePath, seen)
	}
}

func TestRecursiveWatcherWatchesDirectoriesCreatedAfterStart(t *testing.T) {
	root := t.TempDir()
	watcher, err := newRecursiveWatcher(root, nil)
	if err != nil {
		t.Fatalf("newRecursiveWatcher: %v", err)
	}
	defer watcher.Close()

	// A whole subtree appears after the watcher started: the new directories
	// must be watched dynamically so a file created inside them still yields
	// an event.
	nestedDir := filepath.Join(root, "albums", "2026")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	collectEventsUntil(t, watcher, map[string]bool{filepath.Join(root, "albums"): true}, 2*time.Second)

	// Give the dynamic watch a moment to land, then create the file.
	time.Sleep(100 * time.Millisecond)
	deepFile := filepath.Join(nestedDir, "deep.jpg")
	if err := os.WriteFile(deepFile, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	seen := collectEventsUntil(t, watcher, map[string]bool{deepFile: true}, 2*time.Second)
	if !seen[deepFile] {
		t.Fatalf("event for file inside post-start directory not received, saw %v", seen)
	}
}

func TestRecursiveWatcherErrorReportsToFallback(t *testing.T) {
	root := t.TempDir()

	var reported []error
	watcher, err := newRecursiveWatcher(root, func(err error) {
		reported = append(reported, err)
	})
	if err != nil {
		t.Fatalf("newRecursiveWatcher: %v", err)
	}
	defer watcher.Close()

	// Overflow/errors arrive through the error path; nil must be ignored.
	watcher.handleError(nil)
	watcher.handleError(errors.New("queue overflow"))

	if len(reported) != 1 || reported[0].Error() != "queue overflow" {
		t.Fatalf("expected exactly the overflow error reported, got %v", reported)
	}
}
