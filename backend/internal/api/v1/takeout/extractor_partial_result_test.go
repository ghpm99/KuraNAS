package takeout

import (
	"errors"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/libraries"
)

type failingAfterResolver struct {
	imageLibraryPath string
	successfulCalls  int
	callCount        int
	failure          error
}

func (resolver *failingAfterResolver) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	resolver.callCount++
	if resolver.callCount > resolver.successfulCalls {
		return libraries.LibraryDto{}, resolver.failure
	}
	return libraries.LibraryDto{Category: string(category), Path: resolver.imageLibraryPath}, nil
}

func TestExtractTakeoutReturnsAccumulatedResultWhenThirdEntryAborts(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "input.zip")
	createOrderedZip(t, zipPath, []string{"Google Fotos/A.jpg", "Google Fotos/B.jpg", "Google Fotos/C.jpg"})
	libraryFailure := errors.New("library unavailable")
	resolver := &failingAfterResolver{
		imageLibraryPath: filepath.Join(tempDir, "Imagens"),
		successfulCalls:  2,
		failure:          libraryFailure,
	}

	extraction, err := ExtractTakeout(zipPath, resolver)

	if !errors.Is(err, libraryFailure) {
		t.Fatalf("expected library failure, got %v", err)
	}
	if extraction.Imported != 2 {
		t.Fatalf("expected 2 imported entries, got %d", extraction.Imported)
	}
	if len(extraction.Files) != 2 {
		t.Fatalf("expected 2 extracted files, got %d", len(extraction.Files))
	}
	if extraction.TotalEntries != 3 {
		t.Fatalf("expected 3 processed entries, got %d", extraction.TotalEntries)
	}
}

func createOrderedZip(t *testing.T, zipPath string, entryNames []string) {
	t.Helper()
	files := make(map[string]string, len(entryNames))
	for position, entryName := range entryNames {
		files[entryName] = "content-" + string(rune('a'+position))
	}
	createZipFile(t, zipPath, files)
}
