package takeout

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

	if len(result.Files) != 1 || result.Ignored != 4 {
		t.Fatalf("expected 1 extracted and 4 skipped, got %d and %d", len(result.Files), result.Ignored)
	}
	if _, err := os.Stat(filepath.Join(imageLib, "takeout", "ok", "photo.jpg")); err != nil {
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

func extractSingleLibrary(t *testing.T, files map[string]string) (string, ExtractResult, func(zipFiles map[string]string) ExtractResult) {
	t.Helper()
	tempDir := t.TempDir()
	imageLib := filepath.Join(tempDir, "Imagens")
	resolver := &extractorLibraryResolverMock{paths: map[libraries.LibraryCategory]string{
		libraries.LibraryCategoryImages: imageLib,
		libraries.LibraryCategoryVideos: filepath.Join(tempDir, "Videos"),
	}}
	extract := func(zipFiles map[string]string) ExtractResult {
		zipPath := filepath.Join(tempDir, "input.zip")
		createZipFile(t, zipPath, zipFiles)
		extraction, err := ExtractTakeout(zipPath, resolver)
		if err != nil {
			t.Fatalf("ExtractTakeout returned error: %v", err)
		}
		return extraction
	}
	return imageLib, extract(files), extract
}

func readFileContent(t *testing.T, filePath string) string {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected file %s: %v", filePath, err)
	}
	return string(content)
}

func TestExtractTakeoutKeepsSameNameFromDifferentAlbums(t *testing.T) {
	imageLib, extraction, _ := extractSingleLibrary(t, map[string]string{
		"Takeout/Google Fotos/Álbum A/IMG_0001.JPG": "photo-a",
		"Takeout/Google Fotos/Álbum B/IMG_0001.JPG": "photo-b",
	})

	if extraction.Imported != 2 || extraction.TotalEntries != 2 {
		t.Fatalf("expected 2 imported of 2, got %+v", extraction)
	}
	if got := readFileContent(t, filepath.Join(imageLib, "takeout", "Álbum A", "IMG_0001.JPG")); got != "photo-a" {
		t.Fatalf("album A content = %q", got)
	}
	if got := readFileContent(t, filepath.Join(imageLib, "takeout", "Álbum B", "IMG_0001.JPG")); got != "photo-b" {
		t.Fatalf("album B content = %q", got)
	}
}

func TestExtractTakeoutStripsGooglePhotosPrefixCaseInsensitive(t *testing.T) {
	imageLib, extraction, _ := extractSingleLibrary(t, map[string]string{
		"TAKEOUT/Google Photos/Viagem/a.jpg": "a",
		"Takeout/Google Fotos/b.jpg":         "b",
	})

	if extraction.Imported != 2 {
		t.Fatalf("expected 2 imported, got %+v", extraction)
	}
	readFileContent(t, filepath.Join(imageLib, "takeout", "Viagem", "a.jpg"))
	readFileContent(t, filepath.Join(imageLib, "takeout", "b.jpg"))
}

func TestExtractTakeoutSuffixesDifferentContentInSameFolder(t *testing.T) {
	tempDir := t.TempDir()
	imageLib := filepath.Join(tempDir, "Imagens")
	existingPath := filepath.Join(imageLib, "takeout", "Álbum", "IMG_0001.JPG")
	if err := os.MkdirAll(filepath.Dir(existingPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(existingPath, []byte("original"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(existingPath), "IMG_0001 (2).JPG"), []byte("second"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	zipPath := filepath.Join(tempDir, "input.zip")
	createZipFile(t, zipPath, map[string]string{"Álbum/IMG_0001.JPG": "novo"})

	extraction, err := ExtractTakeout(zipPath, &extractorLibraryResolverMock{paths: map[libraries.LibraryCategory]string{
		libraries.LibraryCategoryImages: imageLib,
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if extraction.Imported != 1 || extraction.AlreadyExisting != 0 {
		t.Fatalf("expected 1 imported, got %+v", extraction)
	}
	if got := readFileContent(t, existingPath); got != "original" {
		t.Fatalf("original overwritten: %q", got)
	}
	if got := readFileContent(t, filepath.Join(filepath.Dir(existingPath), "IMG_0001 (3).JPG")); got != "novo" {
		t.Fatalf("suffixed content = %q", got)
	}
}

func TestExtractTakeoutCountsSuffixedIdenticalFileAsAlreadyExisting(t *testing.T) {
	tempDir := t.TempDir()
	imageLib := filepath.Join(tempDir, "Imagens")
	folder := filepath.Join(imageLib, "takeout", "Álbum")
	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for fileName, content := range map[string]string{"IMG_0001.JPG": "original", "IMG_0001 (2).JPG": "novo"} {
		if err := os.WriteFile(filepath.Join(folder, fileName), []byte(content), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	zipPath := filepath.Join(tempDir, "input.zip")
	createZipFile(t, zipPath, map[string]string{"Álbum/IMG_0001.JPG": "novo"})

	extraction, err := ExtractTakeout(zipPath, &extractorLibraryResolverMock{paths: map[libraries.LibraryCategory]string{
		libraries.LibraryCategoryImages: imageLib,
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if extraction.AlreadyExisting != 1 || extraction.Imported != 0 || len(extraction.Files) != 0 {
		t.Fatalf("expected 1 already existing, got %+v", extraction)
	}
}

func TestExtractTakeoutTwiceDoesNotRewriteAnything(t *testing.T) {
	zipFiles := map[string]string{
		"Takeout/Google Fotos/A/IMG_0001.JPG": "photo-a",
		"Takeout/Google Fotos/B/IMG_0001.JPG": "photo-b",
		"Takeout/Google Fotos/B/clip.mp4":     "clip",
	}
	imageLib, first, extractAgain := extractSingleLibrary(t, zipFiles)
	if first.Imported != 3 {
		t.Fatalf("expected 3 imported, got %+v", first)
	}
	photoPath := filepath.Join(imageLib, "takeout", "A", "IMG_0001.JPG")
	pastModTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(photoPath, pastModTime, pastModTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	second := extractAgain(zipFiles)

	if second.Imported != 0 || second.AlreadyExisting != 3 || second.TotalEntries != 3 || len(second.Files) != 0 {
		t.Fatalf("expected all already existing, got %+v", second)
	}
	info, err := os.Stat(photoPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.ModTime().Equal(pastModTime) {
		t.Fatalf("file was rewritten: modtime %v", info.ModTime())
	}
	if _, err := os.Stat(filepath.Join(imageLib, "takeout", "A", "IMG_0001 (2).JPG")); err == nil {
		t.Fatalf("unexpected suffixed duplicate")
	}
}

func TestExtractTakeoutCountsIgnoredAndMetadataSeparately(t *testing.T) {
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"A/IMG_0001.JPG":      "photo",
		"A/IMG_0001.JPG.json": `{"title":"IMG_0001.JPG"}`,
		"A/README.txt":        "text",
	})

	if extraction.TotalEntries != 2 || extraction.Imported != 1 || extraction.Ignored != 1 || extraction.FailedEntries != 0 {
		t.Fatalf("unexpected counts: %+v", extraction)
	}
}

func TestExtractTakeoutMatchesMetadataFromSameFolderOnly(t *testing.T) {
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"Takeout/Google Fotos/A/IMG_0001.JPG":      "photo-a",
		"Takeout/Google Fotos/A/IMG_0001.JPG.json": `{"title":"IMG_0001.JPG","photoTakenTime":{"timestamp":"111"}}`,
		"Takeout/Google Fotos/B/IMG_0001.JPG":      "photo-b",
		"Takeout/Google Fotos/B/IMG_0001.JPG.json": `{"title":"IMG_0001.JPG","photoTakenTime":{"timestamp":"222"}}`,
		"Takeout/Google Fotos/C/IMG_0001.JPG":      "photo-c",
	})

	timestampByFolder := map[string]string{}
	for _, extracted := range extraction.Files {
		folder := filepath.Base(filepath.Dir(extracted.DestinationPath))
		if extracted.Metadata == nil {
			timestampByFolder[folder] = ""
			continue
		}
		timestampByFolder[folder] = extracted.Metadata.PhotoTakenTime.Timestamp
	}
	expected := map[string]string{"A": "111", "B": "222", "C": ""}
	for folder, timestamp := range expected {
		if timestampByFolder[folder] != timestamp {
			t.Fatalf("folder %s timestamp = %q, want %q", folder, timestampByFolder[folder], timestamp)
		}
	}
}

func TestExtractTakeoutMatchesMetadataByTitleInSameFolder(t *testing.T) {
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"A/IMG_0001.JPG":  "photo",
		"A/IMG_0001.json": `{"title":"IMG_0001.JPG","photoTakenTime":{"timestamp":"333"}}`,
	})

	if len(extraction.Files) != 1 || extraction.Files[0].Metadata == nil || extraction.Files[0].Metadata.PhotoTakenTime.Timestamp != "333" {
		t.Fatalf("expected metadata matched by title, got %+v", extraction.Files)
	}
}

func TestExtractTakeoutUnsafeEntriesAreIgnoredAndCounted(t *testing.T) {
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"../x.jpg": "x",
		"ok/y.jpg": "y",
	})

	if extraction.TotalEntries != 2 || extraction.Ignored != 1 || extraction.Imported != 1 {
		t.Fatalf("unexpected counts: %+v", extraction)
	}
	if strings.Contains(extraction.Files[0].DestinationPath, "..") {
		t.Fatalf("unsafe destination %s", extraction.Files[0].DestinationPath)
	}
}

func findExtractedBySourceSuffix(t *testing.T, extraction ExtractResult, sourceSuffix string) ExtractedFile {
	t.Helper()
	for _, extracted := range extraction.Files {
		if strings.HasSuffix(extracted.SourcePath, sourceSuffix) {
			return extracted
		}
	}
	t.Fatalf("no extracted file with source suffix %q in %+v", sourceSuffix, extraction.Files)
	return ExtractedFile{}
}

func assertModTimeIsUnix(t *testing.T, filePath string, unixSeconds int64) {
	t.Helper()
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat %s: %v", filePath, err)
	}
	if !info.ModTime().Equal(time.Unix(unixSeconds, 0)) {
		t.Fatalf("expected mod time %v, got %v", time.Unix(unixSeconds, 0), info.ModTime())
	}
}

func TestExtractTakeoutSetsModTimeFromPhotoTakenTimeInEachFolder(t *testing.T) {
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"Takeout/Google Fotos/A/IMG_0001.JPG":                            "photo-a",
		"Takeout/Google Fotos/A/IMG_0001.JPG.supplemental-metadata.json": `{"title":"IMG_0001.JPG","photoTakenTime":{"timestamp":"1400000000"},"creationTime":{"timestamp":"1999999999"}}`,
		"Takeout/Google Fotos/B/IMG_0001.JPG":                            "photo-b",
		"Takeout/Google Fotos/B/IMG_0001.JPG.json":                       `{"title":"IMG_0001.JPG","photoTakenTime":{"timestamp":"1450000000"}}`,
	})

	assertModTimeIsUnix(t, findExtractedBySourceSuffix(t, extraction, "A/IMG_0001.JPG").DestinationPath, 1400000000)
	assertModTimeIsUnix(t, findExtractedBySourceSuffix(t, extraction, "B/IMG_0001.JPG").DestinationPath, 1450000000)
}

func TestExtractTakeoutSetsModTimeFromCreationTimeWhenPhotoTakenTimeIsMissing(t *testing.T) {
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"A/IMG_0002.JPG":      "photo",
		"A/IMG_0002.JPG.json": `{"title":"IMG_0002.JPG","creationTime":{"timestamp":"1420000000"}}`,
	})

	assertModTimeIsUnix(t, extraction.Files[0].DestinationPath, 1420000000)
}

func TestExtractTakeoutKeepsModTimeWithoutValidDate(t *testing.T) {
	startedAt := time.Now().Add(-time.Minute)
	_, extraction, _ := extractSingleLibrary(t, map[string]string{
		"A/IMG_0003.JPG":      "photo",
		"A/IMG_0003.JPG.json": `{"title":"IMG_0003.JPG","photoTakenTime":{"timestamp":"not-a-number"}}`,
	})

	info, err := os.Stat(extraction.Files[0].DestinationPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.ModTime().Before(startedAt) {
		t.Fatalf("mod time must stay at extraction time, got %v", info.ModTime())
	}
}

func TestExtractTakeoutNeitherImportsNorIgnoresMetadataJSON(t *testing.T) {
	libraryPath, extraction, _ := extractSingleLibrary(t, map[string]string{
		"A/IMG_0004.JPG": "photo",
		"A/IMG_0004.JPG.supplemental-metadata.json": `{"title":"IMG_0004.JPG"}`,
		"A/metadata.json": `{"title":"album"}`,
	})

	if extraction.Imported != 1 || extraction.Ignored != 0 || extraction.TotalEntries != 1 || len(extraction.Files) != 1 {
		t.Fatalf("JSON must be neither imported nor ignored: %+v", extraction)
	}
	jsonFiles, err := filepath.Glob(filepath.Join(libraryPath, "takeout", "A", "*.json"))
	if err != nil || len(jsonFiles) != 0 {
		t.Fatalf("no JSON may be written to the library, got %v (err %v)", jsonFiles, err)
	}
}
