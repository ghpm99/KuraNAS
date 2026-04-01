package watcher

import (
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/watchfolders"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanWatchFolderFiltersByExtensionAndMtime(t *testing.T) {
	root := t.TempDir()
	image := filepath.Join(root, "photo.jpg")
	music := filepath.Join(root, "song.mp3")
	unknown := filepath.Join(root, "file.bin")
	if err := os.WriteFile(image, []byte("x"), 0644); err != nil {
		t.Fatalf("write image: %v", err)
	}
	if err := os.WriteFile(music, []byte("x"), 0644); err != nil {
		t.Fatalf("write music: %v", err)
	}
	if err := os.WriteFile(unknown, []byte("x"), 0644); err != nil {
		t.Fatalf("write unknown: %v", err)
	}

	lastScan := time.Now().Add(-1 * time.Minute)
	files, err := ScanWatchFolder(watchfolders.WatchFolderModel{Path: root, LastScanAt: &lastScan, Enabled: true})
	if err != nil {
		t.Fatalf("ScanWatchFolder returned error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	categories := map[libraries.LibraryCategory]bool{}
	for _, file := range files {
		categories[file.Category] = true
	}
	if !categories[libraries.LibraryCategoryImages] || !categories[libraries.LibraryCategoryMusic] {
		t.Fatalf("unexpected categories: %+v", categories)
	}
}
