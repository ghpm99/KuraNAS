package watcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveToLibraryMovesAndRenamesOnConflict(t *testing.T) {
	sourceDir := t.TempDir()
	libraryDir := t.TempDir()

	sourcePath := filepath.Join(sourceDir, "example.jpg")
	if err := os.WriteFile(sourcePath, []byte("payload"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(libraryDir, "example.jpg"), []byte("existing"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	movedPath, err := MoveToLibrary(ScannedFile{SourcePath: sourcePath}, libraryDir)
	if err != nil {
		t.Fatalf("MoveToLibrary returned error: %v", err)
	}

	if movedPath == filepath.Join(libraryDir, "example.jpg") {
		t.Fatalf("expected conflict suffix in destination path, got %s", movedPath)
	}
	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("expected moved file to exist: %v", err)
	}
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Fatalf("expected source file to be removed")
	}
}
