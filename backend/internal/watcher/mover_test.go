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

func TestMoveToLibraryCreatesMissingLibraryDir(t *testing.T) {
	sourceDir := t.TempDir()
	// Biblioteca aponta para um diretório que ainda não existe: MoveToLibrary deve criá-lo.
	libraryDir := filepath.Join(t.TempDir(), "images", "library")

	sourcePath := filepath.Join(sourceDir, "photo.jpg")
	if err := os.WriteFile(sourcePath, []byte("payload"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	movedPath, err := MoveToLibrary(ScannedFile{SourcePath: sourcePath}, libraryDir)
	if err != nil {
		t.Fatalf("MoveToLibrary returned error: %v", err)
	}
	if movedPath != filepath.Join(libraryDir, "photo.jpg") {
		t.Fatalf("expected destination without suffix, got %s", movedPath)
	}
	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("expected moved file to exist: %v", err)
	}
}

func TestResolveFileConflictWalksUntilFreeName(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "doc.txt")
	// doc.txt e doc_1.txt já existem -> deve resolver para doc_2.txt.
	for _, name := range []string{"doc.txt", "doc_1.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	resolved, err := resolveFileConflict(target)
	if err != nil {
		t.Fatalf("resolveFileConflict returned error: %v", err)
	}
	if resolved != filepath.Join(dir, "doc_2.txt") {
		t.Fatalf("expected doc_2.txt, got %s", resolved)
	}
}
