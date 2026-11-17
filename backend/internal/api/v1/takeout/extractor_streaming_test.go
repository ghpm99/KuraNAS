package takeout

import (
	"archive/zip"
	"bytes"
	"hash/crc32"
	"io"
	"nas-go/api/internal/api/v1/libraries"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	largeEntryBytes          = 200 << 20
	maxExtractionAllocations = 32 << 20
)

type extractionFixture struct {
	zipPath   string
	imageLib  string
	resolver  *extractorLibraryResolverMock
	extracted ExtractResult
}

func newExtractionFixture(t *testing.T) *extractionFixture {
	t.Helper()
	tempDir := t.TempDir()
	imageLib := filepath.Join(tempDir, "Imagens")
	return &extractionFixture{
		zipPath:  filepath.Join(tempDir, "input.zip"),
		imageLib: imageLib,
		resolver: &extractorLibraryResolverMock{paths: map[libraries.LibraryCategory]string{
			libraries.LibraryCategoryImages: imageLib,
			libraries.LibraryCategoryVideos: filepath.Join(tempDir, "Videos"),
		}},
	}
}

func writeStoredEntry(t *testing.T, writer *zip.Writer, name string, content io.Reader) {
	t.Helper()
	entryWriter, err := writer.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
	if err != nil {
		t.Fatalf("create zip entry %s: %v", name, err)
	}
	if _, err := io.Copy(entryWriter, content); err != nil {
		t.Fatalf("write zip entry %s: %v", name, err)
	}
}

func writeCorruptStoredEntry(t *testing.T, writer *zip.Writer, name string, content []byte) {
	t.Helper()
	entryWriter, err := writer.CreateRaw(&zip.FileHeader{
		Name:               name,
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE(content) + 1,
		CompressedSize64:   uint64(len(content)),
		UncompressedSize64: uint64(len(content)),
	})
	if err != nil {
		t.Fatalf("create corrupt zip entry %s: %v", name, err)
	}
	if _, err := entryWriter.Write(content); err != nil {
		t.Fatalf("write corrupt zip entry %s: %v", name, err)
	}
}

func buildZip(t *testing.T, zipPath string, fill func(writer *zip.Writer)) {
	t.Helper()
	target, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer target.Close()
	writer := zip.NewWriter(target)
	fill(writer)
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

type repeatingReader struct {
	remaining int64
	block     []byte
}

func (r *repeatingReader) Read(buffer []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	count := min(int64(len(buffer)), r.remaining, int64(len(r.block)))
	copy(buffer[:count], r.block[:count])
	r.remaining -= count
	return int(count), nil
}

func listPartialFiles(t *testing.T, root string) []string {
	t.Helper()
	partialFiles := make([]string, 0)
	err := filepath.WalkDir(root, func(entryPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if strings.HasSuffix(entryPath, ".part") {
			partialFiles = append(partialFiles, entryPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return partialFiles
}

func TestExtractTakeoutStreamsLargeEntryWithBoundedMemory(t *testing.T) {
	fixture := newExtractionFixture(t)
	buildZip(t, fixture.zipPath, func(writer *zip.Writer) {
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/big.jpg", &repeatingReader{
			remaining: largeEntryBytes,
			block:     bytes.Repeat([]byte("kuranas-takeout-"), 1<<16),
		})
	})

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	extraction, err := ExtractTakeout(fixture.zipPath, fixture.resolver)
	runtime.ReadMemStats(&after)
	if err != nil {
		t.Fatalf("ExtractTakeout returned error: %v", err)
	}

	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("allocated %d bytes extracting a %d byte entry", allocated, largeEntryBytes)
	if allocated >= maxExtractionAllocations {
		t.Fatalf("extraction allocated %d bytes, limit is %d", allocated, maxExtractionAllocations)
	}
	if extraction.Imported != 1 || extraction.FailedEntries != 0 {
		t.Fatalf("unexpected result: %+v", extraction)
	}
	extractedInfo, statErr := os.Stat(filepath.Join(fixture.imageLib, "takeout", "Album", "big.jpg"))
	if statErr != nil || extractedInfo.Size() != largeEntryBytes {
		t.Fatalf("expected extracted file of %d bytes, got %v (%v)", largeEntryBytes, extractedInfo, statErr)
	}
	if partials := listPartialFiles(t, fixture.imageLib); len(partials) != 0 {
		t.Fatalf("partial files left behind: %v", partials)
	}
}

func TestExtractTakeoutCorruptEntryIsFailedAndNextEntriesAreExtracted(t *testing.T) {
	fixture := newExtractionFixture(t)
	buildZip(t, fixture.zipPath, func(writer *zip.Writer) {
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/before.jpg", strings.NewReader("before"))
		writeCorruptStoredEntry(t, writer, "Takeout/Google Fotos/Album/broken.jpg", []byte("broken content"))
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/after.jpg", strings.NewReader("after"))
	})

	extraction, err := ExtractTakeout(fixture.zipPath, fixture.resolver)
	if err != nil {
		t.Fatalf("ExtractTakeout returned error: %v", err)
	}

	if extraction.FailedEntries != 1 || extraction.Imported != 2 || extraction.TotalEntries != 3 {
		t.Fatalf("unexpected counters: %+v", extraction)
	}
	albumDir := filepath.Join(fixture.imageLib, "takeout", "Album")
	if readFileContent(t, filepath.Join(albumDir, "before.jpg")) != "before" ||
		readFileContent(t, filepath.Join(albumDir, "after.jpg")) != "after" {
		t.Fatalf("entries around the corrupt one were not extracted")
	}
	if _, statErr := os.Stat(filepath.Join(albumDir, "broken.jpg")); !os.IsNotExist(statErr) {
		t.Fatalf("corrupt entry must not be left at its destination: %v", statErr)
	}
	if partials := listPartialFiles(t, fixture.imageLib); len(partials) != 0 {
		t.Fatalf("partial files left behind: %v", partials)
	}
}

func TestExtractTakeoutIgnoresMetadataJSONAboveOneMegabyte(t *testing.T) {
	fixture := newExtractionFixture(t)
	oversizedJSON := `{"title":"big.jpg","photoTakenTime":{"timestamp":"1700000000"},"padding":"` +
		strings.Repeat("x", maxMetadataEntryBytes) + `"}`
	buildZip(t, fixture.zipPath, func(writer *zip.Writer) {
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/big.jpg", strings.NewReader("image"))
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/big.jpg.json", strings.NewReader(oversizedJSON))
	})

	extraction, err := ExtractTakeout(fixture.zipPath, fixture.resolver)
	if err != nil {
		t.Fatalf("ExtractTakeout returned error: %v", err)
	}

	if extraction.Imported != 1 || extraction.FailedEntries != 0 {
		t.Fatalf("unexpected result: %+v", extraction)
	}
	if extraction.Files[0].Metadata != nil {
		t.Fatalf("oversized metadata must be ignored, got %+v", extraction.Files[0].Metadata)
	}
}

type progressCall struct {
	processedEntries int
	totalEntries     int
}

func TestExtractTakeoutWithProgressReportsEveryCountedEntryEndingAtTotal(t *testing.T) {
	fixture := newExtractionFixture(t)
	buildZip(t, fixture.zipPath, func(writer *zip.Writer) {
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/a.jpg", strings.NewReader("a"))
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/a.jpg.json", strings.NewReader(`{"title":"a.jpg"}`))
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/notes.txt", strings.NewReader("notes"))
		writeCorruptStoredEntry(t, writer, "Takeout/Google Fotos/Album/b.jpg", []byte("broken"))
		writeStoredEntry(t, writer, "Takeout/Google Fotos/Album/c.jpg", strings.NewReader("c"))
	})

	calls := make([]progressCall, 0)
	extraction, err := ExtractTakeoutWithProgress(fixture.zipPath, fixture.resolver, func(processedEntries int, totalEntries int) {
		calls = append(calls, progressCall{processedEntries, totalEntries})
	})
	if err != nil {
		t.Fatalf("ExtractTakeoutWithProgress returned error: %v", err)
	}

	expectedCalls := []progressCall{{1, 4}, {2, 4}, {3, 4}, {4, 4}}
	if len(calls) != len(expectedCalls) {
		t.Fatalf("expected calls %v, got %v", expectedCalls, calls)
	}
	for index, expected := range expectedCalls {
		if calls[index] != expected {
			t.Fatalf("expected calls %v, got %v", expectedCalls, calls)
		}
	}
	if extraction.TotalEntries != 4 {
		t.Fatalf("expected TotalEntries 4, got %d", extraction.TotalEntries)
	}
}
