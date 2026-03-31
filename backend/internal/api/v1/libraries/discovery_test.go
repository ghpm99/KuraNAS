package libraries

import (
	"os"
	"path/filepath"
	"testing"
)

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func TestFindSlugMatch_ExactMatch(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "Imagens"))

	result := findSlugMatch(root, categorySlugs[LibraryCategoryImages])
	if result != filepath.Join(root, "Imagens") {
		t.Fatalf("expected Imagens folder, got %s", result)
	}
}

func TestFindSlugMatch_CaseInsensitive(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "dOcUmEnToS"))

	result := findSlugMatch(root, categorySlugs[LibraryCategoryDocuments])
	if result != filepath.Join(root, "dOcUmEnToS") {
		t.Fatalf("expected case-insensitive match, got %s", result)
	}
}

func TestFindSlugMatch_NoMatch(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "random"))

	result := findSlugMatch(root, categorySlugs[LibraryCategoryMusic])
	if result != "" {
		t.Fatalf("expected empty result, got %s", result)
	}
}

func TestFindBestMatchByFileCount(t *testing.T) {
	root := t.TempDir()
	fav := filepath.Join(root, "Fav")
	other := filepath.Join(root, "Other")
	mustMkdir(t, fav)
	mustMkdir(t, other)
	mustWriteFile(t, filepath.Join(fav, "a.jpg"))
	mustWriteFile(t, filepath.Join(fav, "b.png"))
	mustWriteFile(t, filepath.Join(other, "c.jpg"))

	result := findBestMatchByFileCount(root, categoryExtensions[LibraryCategoryImages])
	if result != fav {
		t.Fatalf("expected %s, got %s", fav, result)
	}
}

func TestFindBestMatchByFileCount_NoDirs(t *testing.T) {
	root := t.TempDir()
	result := findBestMatchByFileCount(root, categoryExtensions[LibraryCategoryVideos])
	if result != "" {
		t.Fatalf("expected empty result, got %s", result)
	}
}

func TestResolveLibraryPath_SlugFound(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "Videos"))
	mustMkdir(t, filepath.Join(root, "Else"))

	result := resolveLibraryPath(root, LibraryCategoryVideos)
	if result != filepath.Join(root, "Videos") {
		t.Fatalf("expected slug match to win, got %s", result)
	}
}

func TestResolveLibraryPath_FallbackToFileCount(t *testing.T) {
	root := t.TempDir()
	best := filepath.Join(root, "MoviesA")
	other := filepath.Join(root, "MoviesB")
	mustMkdir(t, best)
	mustMkdir(t, other)
	mustWriteFile(t, filepath.Join(best, "a.mp4"))
	mustWriteFile(t, filepath.Join(best, "b.mkv"))
	mustWriteFile(t, filepath.Join(other, "c.mp4"))

	result := resolveLibraryPath(root, LibraryCategoryVideos)
	if result != best {
		t.Fatalf("expected file-count fallback %s, got %s", best, result)
	}
}

func TestResolveLibraryPath_CreatesFolder(t *testing.T) {
	root := t.TempDir()
	result := resolveLibraryPath(root, LibraryCategoryDocuments)
	expected := filepath.Join(root, "Documentos")
	if result != expected {
		t.Fatalf("expected default folder %s, got %s", expected, result)
	}
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected folder to be created: %v", err)
	}
}
