package takeout

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const maxMetadataEntryBytes = 1 << 20

// ProgressReporter receives the entries processed so far and the total of
// entries the extraction will process.
type ProgressReporter func(processedEntries int, totalEntries int)

func parseTakeoutMetadata(jsonBytes []byte) (TakeoutMetadata, error) {
	var metadata TakeoutMetadata
	if err := json.Unmarshal(jsonBytes, &metadata); err != nil {
		return TakeoutMetadata{}, err
	}
	return metadata, nil
}

func classifyFile(fileName string, mimeType string) libraries.LibraryCategory {
	ext := strings.ToLower(filepath.Ext(fileName))
	if utils.GetFormatTypeByExtension(ext).Type == utils.FormatTypeImage || ext == ".raw" {
		return libraries.LibraryCategoryImages
	}
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".ts":
		return libraries.LibraryCategoryVideos
	}

	if strings.HasPrefix(mimeType, "image/") {
		return libraries.LibraryCategoryImages
	}
	if strings.HasPrefix(mimeType, "video/") {
		return libraries.LibraryCategoryVideos
	}

	return ""
}

func buildDestinationPath(libraryPath string, relativeEntryPath string) (string, error) {
	extractionRoot := filepath.Join(libraryPath, "takeout")
	return safeExtractionPath(extractionRoot, sanitizeEntryPath(relativeEntryPath))
}

func safeExtractionPath(extractionRoot string, entryName string) (string, error) {
	slashedName := strings.ReplaceAll(entryName, "\\", "/")
	if hasDriveLetterPrefix(slashedName) || filepath.VolumeName(slashedName) != "" {
		return "", ErrUnsafeArchiveEntry
	}
	if strings.HasPrefix(slashedName, "/") {
		return "", ErrUnsafeArchiveEntry
	}

	target := filepath.Join(extractionRoot, filepath.FromSlash(slashedName))
	relativeToRoot, err := filepath.Rel(extractionRoot, target)
	if err != nil || filepath.IsAbs(relativeToRoot) {
		return "", ErrUnsafeArchiveEntry
	}
	if relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", ErrUnsafeArchiveEntry
	}
	return target, nil
}

func hasDriveLetterPrefix(entryName string) bool {
	if len(entryName) < 2 || entryName[1] != ':' {
		return false
	}
	firstChar := entryName[0]
	return (firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')
}

func isUnsafeArchiveEntry(entry *zip.File) bool {
	if entry.Mode()&os.ModeSymlink != 0 {
		applog.Warn("takeout: skipping symlink entry", "entry", entry.Name)
		return true
	}
	if _, err := safeExtractionPath(".", entry.Name); err != nil {
		applog.Warn("takeout: skipping entry with unsafe path", "entry", entry.Name)
		return true
	}
	return false
}

func isCountedEntry(entry *zip.File) bool {
	if entry.FileInfo().IsDir() {
		return false
	}
	if entry.Mode()&os.ModeSymlink != 0 {
		return true
	}
	if _, err := safeExtractionPath(".", entry.Name); err != nil {
		return true
	}
	return !isMetadataFileName(stripGooglePhotosPrefix(entry.Name))
}

func countCountedEntries(entries []*zip.File) int {
	counted := 0
	for _, entry := range entries {
		if isCountedEntry(entry) {
			counted++
		}
	}
	return counted
}

// ExtractTakeout extracts the photos and videos of a Google Takeout zip into
// the matching libraries under takeout/<folders of the entry>/, never
// overwriting existing files. Unreadable entries are counted as failed and the
// extraction continues; an invalid zip or an unresolvable library aborts it.
func ExtractTakeout(zipPath string, libraryResolver LibraryResolverInterface) (ExtractResult, error) {
	return ExtractTakeoutWithProgress(zipPath, libraryResolver, nil)
}

// ExtractTakeoutWithProgress behaves like ExtractTakeout and calls reportProgress
// after every processed entry with the number of entries processed so far and
// the total of entries that will be processed. A nil reportProgress is allowed.
// When an error aborts the extraction after it started, the returned result
// still reports everything extracted up to that point.
func ExtractTakeoutWithProgress(zipPath string, libraryResolver LibraryResolverInterface, reportProgress ProgressReporter) (ExtractResult, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return ExtractResult{}, ErrInvalidZipFile
	}
	defer reader.Close()

	metadataIndex := indexMetadataEntries(reader.File)
	extraction := ExtractResult{
		Files: make([]ExtractedFile, 0),
	}
	totalEntries := countCountedEntries(reader.File)

	for _, entry := range reader.File {
		if !isCountedEntry(entry) {
			continue
		}
		if err := extractCountedEntry(entry, libraryResolver, metadataIndex, &extraction); err != nil {
			return extraction, err
		}
		if reportProgress != nil {
			reportProgress(extraction.TotalEntries, totalEntries)
		}
	}

	return extraction, nil
}

func extractCountedEntry(entry *zip.File, libraryResolver LibraryResolverInterface, metadataLookup *metadataIndex, extraction *ExtractResult) error {
	extraction.TotalEntries++

	if isUnsafeArchiveEntry(entry) {
		extraction.Ignored++
		return nil
	}

	relativePath := stripGooglePhotosPrefix(entry.Name)
	entryName := path.Base(relativePath)
	category := classifyFile(entryName, mime.TypeByExtension(strings.ToLower(filepath.Ext(entryName))))
	if category == "" {
		extraction.Ignored++
		return nil
	}

	libraryDto, resolveErr := libraryResolver.GetLibraryByCategory(category)
	if resolveErr != nil {
		return fmt.Errorf("resolve library %s: %w", category, resolveErr)
	}

	destinationPath, pathErr := buildDestinationPath(libraryDto.Path, relativePath)
	if pathErr != nil {
		applog.Warn("takeout: skipping entry with unsafe destination", "entry", entry.Name)
		extraction.Ignored++
		return nil
	}

	writtenPath, isAlreadyExisting, writeErr := extractEntryWithoutOverwrite(entry, destinationPath)
	if writeErr != nil {
		applog.Warn("takeout: failed to extract entry", "entry", entry.Name, "error", writeErr)
		extraction.FailedEntries++
		return nil
	}
	if isAlreadyExisting {
		extraction.AlreadyExisting++
		return nil
	}
	extraction.Imported++

	metadata := metadataLookup.find(relativePath)
	applyTakenTimeToFile(writtenPath, metadata)
	extraction.Files = append(extraction.Files, ExtractedFile{
		SourcePath:      entry.Name,
		DestinationPath: writtenPath,
		Category:        string(category),
		Metadata:        metadata,
	})
	return nil
}

func isMetadataFileName(entryPath string) bool {
	return strings.HasSuffix(strings.ToLower(entryPath), ".json")
}

func indexMetadataEntries(entries []*zip.File) *metadataIndex {
	index := newMetadataIndex()
	for _, entry := range entries {
		if entry.FileInfo().IsDir() || isUnsafeArchiveEntry(entry) {
			continue
		}
		relativePath := stripGooglePhotosPrefix(entry.Name)
		if !isMetadataFileName(relativePath) {
			continue
		}
		content, readErr := readMetadataEntry(entry)
		if readErr != nil {
			applog.Warn("takeout: ignoring unreadable metadata entry", "entry", entry.Name, "error", readErr)
			continue
		}
		metadata, parseErr := parseTakeoutMetadata(content)
		if parseErr != nil {
			continue
		}
		index.add(relativePath, metadata)
	}
	return index
}

func readMetadataEntry(entry *zip.File) ([]byte, error) {
	if entry.UncompressedSize64 > maxMetadataEntryBytes {
		return nil, ErrMetadataEntryTooLarge
	}
	reader, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(io.LimitReader(reader, maxMetadataEntryBytes+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxMetadataEntryBytes {
		return nil, ErrMetadataEntryTooLarge
	}
	return content, nil
}
