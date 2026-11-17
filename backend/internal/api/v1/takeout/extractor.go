package takeout

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
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

// ExtractTakeout extracts the photos and videos of a Google Takeout zip into
// the matching libraries under takeout/<folders of the entry>/, never
// overwriting existing files. Unreadable entries are counted as failed and the
// extraction continues; an invalid zip or an unresolvable library aborts it.
func ExtractTakeout(zipPath string, libraryResolver LibraryResolverInterface) (ExtractResult, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return ExtractResult{}, ErrInvalidZipFile
	}
	defer reader.Close()

	metadataByEntryKey := indexMetadataByEntryKey(reader.File)
	extraction := ExtractResult{
		Files: make([]ExtractedFile, 0),
	}

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}

		if isUnsafeArchiveEntry(entry) {
			extraction.TotalEntries++
			extraction.Ignored++
			continue
		}

		relativePath := stripGooglePhotosPrefix(entry.Name)
		if isMetadataFileName(relativePath) {
			continue
		}
		extraction.TotalEntries++

		entryName := path.Base(relativePath)
		category := classifyFile(entryName, mime.TypeByExtension(strings.ToLower(filepath.Ext(entryName))))
		if category == "" {
			extraction.Ignored++
			continue
		}

		libraryDto, resolveErr := libraryResolver.GetLibraryByCategory(category)
		if resolveErr != nil {
			return ExtractResult{}, fmt.Errorf("resolve library %s: %w", category, resolveErr)
		}

		destinationPath, pathErr := buildDestinationPath(libraryDto.Path, relativePath)
		if pathErr != nil {
			applog.Warn("takeout: skipping entry with unsafe destination", "entry", entry.Name)
			extraction.Ignored++
			continue
		}

		writtenPath, isAlreadyExisting, writeErr := extractEntryWithoutOverwrite(entry, destinationPath)
		if writeErr != nil {
			applog.Warn("takeout: failed to extract entry", "entry", entry.Name, "error", writeErr)
			extraction.FailedEntries++
			continue
		}
		if isAlreadyExisting {
			extraction.AlreadyExisting++
			continue
		}
		extraction.Imported++

		extraction.Files = append(extraction.Files, ExtractedFile{
			SourcePath:      entry.Name,
			DestinationPath: writtenPath,
			Category:        string(category),
			Metadata:        findMetadata(metadataByEntryKey, relativePath),
		})
	}

	return extraction, nil
}

func isMetadataFileName(entryPath string) bool {
	return strings.HasSuffix(strings.ToLower(entryPath), ".json")
}

func indexMetadataByEntryKey(entries []*zip.File) map[string]TakeoutMetadata {
	metadataByEntryKey := map[string]TakeoutMetadata{}
	for _, entry := range entries {
		if entry.FileInfo().IsDir() || isUnsafeArchiveEntry(entry) {
			continue
		}
		relativePath := stripGooglePhotosPrefix(entry.Name)
		if !isMetadataFileName(relativePath) {
			continue
		}
		content, readErr := readZipEntry(entry)
		if readErr != nil {
			continue
		}
		metadata, parseErr := parseTakeoutMetadata(content)
		if parseErr != nil {
			continue
		}
		folder := path.Dir(relativePath)
		mediaName := strings.TrimSuffix(path.Base(relativePath), path.Ext(relativePath))
		metadataByEntryKey[metadataEntryKey(folder, mediaName)] = metadata
		if strings.TrimSpace(metadata.Title) != "" {
			metadataByEntryKey[metadataEntryKey(folder, metadata.Title)] = metadata
		}
	}
	return metadataByEntryKey
}

func findMetadata(metadataByEntryKey map[string]TakeoutMetadata, relativePath string) *TakeoutMetadata {
	metadata, found := metadataByEntryKey[metadataEntryKey(path.Dir(relativePath), path.Base(relativePath))]
	if !found {
		return nil
	}
	return &metadata
}

func metadataEntryKey(folder string, mediaName string) string {
	return folder + "/" + mediaName
}

func readZipEntry(entry *zip.File) ([]byte, error) {
	reader, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		if errors.Is(err, io.EOF) {
			return buffer.Bytes(), nil
		}
		return nil, err
	}

	return buffer.Bytes(), nil
}
