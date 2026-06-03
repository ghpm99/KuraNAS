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

func TestClassifyWatchFileByCategory(t *testing.T) {
	cases := []struct {
		path     string
		expected libraries.LibraryCategory
		ok       bool
	}{
		{"a.jpg", libraries.LibraryCategoryImages, true},
		{"a.PNG", libraries.LibraryCategoryImages, true},
		{"a.mp3", libraries.LibraryCategoryMusic, true},
		{"a.mp4", libraries.LibraryCategoryVideos, true},
		{"a.pdf", libraries.LibraryCategoryDocuments, true},
		{"a.bin", "", false},
		{"noext", "", false},
	}
	for _, c := range cases {
		got, ok := classifyWatchFile(c.path)
		if ok != c.ok || got != c.expected {
			t.Fatalf("classifyWatchFile(%q) = (%q,%v), want (%q,%v)", c.path, got, ok, c.expected, c.ok)
		}
	}
}

func TestScanWatchFolderWithoutThresholdRecursesAndSorts(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "nested")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	// Sem LastScanAt o threshold é zero: todos os arquivos de mídia entram, inclusive em subpastas.
	for _, p := range []string{filepath.Join(root, "b.mp4"), filepath.Join(root, "a.jpg"), filepath.Join(sub, "c.mp3")} {
		if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	files, err := ScanWatchFolder(watchfolders.WatchFolderModel{Path: root, Enabled: true})
	if err != nil {
		t.Fatalf("ScanWatchFolder returned error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files (recursive), got %d", len(files))
	}
	// Resultado deve vir ordenado por caminho.
	for i := 1; i < len(files); i++ {
		if files[i-1].SourcePath > files[i].SourcePath {
			t.Fatalf("expected sorted by path, got %s before %s", files[i-1].SourcePath, files[i].SourcePath)
		}
	}
}
