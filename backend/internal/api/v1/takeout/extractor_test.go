package takeout

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/libraries"
)

type extractorLibraryResolverMock struct {
	paths map[libraries.LibraryCategory]string
}

func (m *extractorLibraryResolverMock) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{
		Category: string(category),
		Path:     m.paths[category],
	}, nil
}

func createZipFile(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()
	target, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer target.Close()

	writer := zip.NewWriter(target)
	for name, content := range files {
		entry, createErr := writer.Create(name)
		if createErr != nil {
			t.Fatalf("failed to create zip entry: %v", createErr)
		}
		if _, writeErr := entry.Write([]byte(content)); writeErr != nil {
			t.Fatalf("failed to write zip entry: %v", writeErr)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}

func TestParseTakeoutMetadata(t *testing.T) {
	valid := []byte(`{"title":"IMG_01.jpg","photoTakenTime":{"timestamp":"1","formatted":"x"}}`)
	if _, err := parseTakeoutMetadata(valid); err != nil {
		t.Fatalf("expected valid metadata parse, got %v", err)
	}

	partial := []byte(`{"title":"IMG_02.jpg"}`)
	if _, err := parseTakeoutMetadata(partial); err != nil {
		t.Fatalf("expected partial metadata parse, got %v", err)
	}

	invalid := []byte(`{`)
	if _, err := parseTakeoutMetadata(invalid); err == nil {
		t.Fatalf("expected parse error for invalid metadata")
	}
}

func TestClassifyFile(t *testing.T) {
	if classifyFile("photo.jpg", "") != libraries.LibraryCategoryImages {
		t.Fatalf("expected jpg to map to images")
	}
	if classifyFile("movie.mp4", "") != libraries.LibraryCategoryVideos {
		t.Fatalf("expected mp4 to map to videos")
	}
	if classifyFile("unknown.bin", "application/octet-stream") != "" {
		t.Fatalf("expected unknown file to be ignored")
	}
}

func TestBuildDestinationPath(t *testing.T) {
	got, err := buildDestinationPath("/data/Imagens", "IMG_01.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join("/data/Imagens", "takeout", "IMG_01.jpg")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestExtractTakeout(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "input.zip")
	imageLib := filepath.Join(tempDir, "Imagens")
	videoLib := filepath.Join(tempDir, "Videos")
	if err := os.MkdirAll(imageLib, 0755); err != nil {
		t.Fatalf("failed to create image lib: %v", err)
	}
	if err := os.MkdirAll(videoLib, 0755); err != nil {
		t.Fatalf("failed to create video lib: %v", err)
	}

	createZipFile(t, zipPath, map[string]string{
		"Google Fotos/IMG_01.jpg":      "img",
		"Google Fotos/IMG_01.jpg.json": `{"title":"IMG_01.jpg","photoTakenTime":{"timestamp":"1705312822"}}`,
		"Google Fotos/VID_01.mp4":      "vid",
		"Google Fotos/README.txt":      "ignored",
	})

	result, err := ExtractTakeout(zipPath, &extractorLibraryResolverMock{
		paths: map[libraries.LibraryCategory]string{
			libraries.LibraryCategoryImages: imageLib,
			libraries.LibraryCategoryVideos: videoLib,
		},
	})
	if err != nil {
		t.Fatalf("ExtractTakeout returned error: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("expected 2 extracted files, got %d", len(result.Files))
	}
}

func TestExtractTakeoutInvalidZip(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "invalid.zip")
	if err := os.WriteFile(zipPath, []byte("not a zip"), 0644); err != nil {
		t.Fatalf("failed to write invalid zip: %v", err)
	}

	_, err := ExtractTakeout(zipPath, &extractorLibraryResolverMock{
		paths: map[libraries.LibraryCategory]string{},
	})
	if err == nil {
		t.Fatalf("expected invalid zip error")
	}
}

func TestSafeExtractionPathRejectsUnsafeEntryNames(t *testing.T) {
	extractionRoot := filepath.Join(t.TempDir(), "takeout")
	unsafeEntryNames := []string{
		"../evil",
		"a/../../evil",
		"/abs",
		"C:\\evil",
		"c:/evil",
		"..\\evil",
		"a\\..\\..\\evil",
		"..",
	}

	for _, entryName := range unsafeEntryNames {
		if _, err := safeExtractionPath(extractionRoot, entryName); !errors.Is(err, ErrUnsafeArchiveEntry) {
			t.Fatalf("expected ErrUnsafeArchiveEntry for %q, got %v", entryName, err)
		}
	}
}

func TestSafeExtractionPathAcceptsNestedEntryName(t *testing.T) {
	extractionRoot := filepath.Join(t.TempDir(), "takeout")

	got, err := safeExtractionPath(extractionRoot, "Google Photos/2020/IMG_01.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(extractionRoot, "Google Photos", "2020", "IMG_01.jpg")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestExtractTakeoutSkipsUnsafeAndSymlinkEntries(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "input.zip")
	imageLib := filepath.Join(tempDir, "Imagens")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	writer := zip.NewWriter(zipFile)
	for _, entryName := range []string{"../evil.jpg", "a/../../evil2.jpg", "C:\\evil3.jpg", "ok/photo.jpg"} {
		entryWriter, createErr := writer.Create(entryName)
		if createErr != nil {
			t.Fatalf("failed to create entry: %v", createErr)
		}
		if _, writeErr := entryWriter.Write([]byte("x")); writeErr != nil {
			t.Fatalf("failed to write entry: %v", writeErr)
		}
	}
	symlinkHeader := &zip.FileHeader{Name: "link.jpg", Method: zip.Store}
	symlinkHeader.SetMode(os.ModeSymlink | 0777)
	symlinkWriter, err := writer.CreateHeader(symlinkHeader)
	if err != nil {
		t.Fatalf("failed to create symlink entry: %v", err)
	}
	if _, err := symlinkWriter.Write([]byte("/etc/passwd")); err != nil {
		t.Fatalf("failed to write symlink entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	zipFile.Close()

	resolver := &extractorLibraryResolverMock{paths: map[libraries.LibraryCategory]string{
		libraries.LibraryCategoryImages: imageLib,
	}}
	result, err := ExtractTakeout(zipPath, resolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 1 || result.SkippedEntries != 4 {
		t.Fatalf("expected 1 extracted and 4 skipped, got %d and %d", len(result.Files), result.SkippedEntries)
	}
	if _, err := os.Stat(filepath.Join(imageLib, "takeout", "photo.jpg")); err != nil {
		t.Fatalf("expected safe entry extracted: %v", err)
	}
	for _, escapedPath := range []string{
		filepath.Join(tempDir, "evil.jpg"),
		filepath.Join(imageLib, "evil2.jpg"),
		filepath.Join(imageLib, "takeout", "link.jpg"),
	} {
		if _, err := os.Stat(escapedPath); err == nil {
			t.Fatalf("unexpected file written at %s", escapedPath)
		}
	}
}
