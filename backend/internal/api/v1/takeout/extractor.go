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

func buildDestinationPath(libraryPath string, fileName string) (string, error) {
	extractionRoot := filepath.Join(libraryPath, "takeout")
	return safeExtractionPath(extractionRoot, sanitizeTakeoutFileName(filepath.Base(fileName)))
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

func ExtractTakeout(zipPath string, libraryResolver LibraryResolverInterface) (ExtractResult, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return ExtractResult{}, ErrInvalidZipFile
	}
	defer reader.Close()

	metadataByName := map[string]TakeoutMetadata{}
	result := ExtractResult{
		Files: make([]ExtractedFile, 0),
	}

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}

		entryName := filepath.Base(entry.Name)
		if strings.HasSuffix(strings.ToLower(entryName), ".json") {
			content, readErr := readZipEntry(entry)
			if readErr != nil {
				continue
			}
			metadata, parseErr := parseTakeoutMetadata(content)
			if parseErr != nil {
				continue
			}
			metadataByName[strings.TrimSuffix(entryName, ".json")] = metadata
			if strings.TrimSpace(metadata.Title) != "" {
				metadataByName[metadata.Title] = metadata
			}
		}
	}

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}

		if isUnsafeArchiveEntry(entry) {
			result.SkippedEntries++
			continue
		}

		entryName := filepath.Base(entry.Name)
		lowerName := strings.ToLower(entryName)
		if strings.HasSuffix(lowerName, ".json") {
			continue
		}

		mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(entryName)))
		category := classifyFile(entryName, mimeType)
		if category == "" {
			continue
		}

		libraryDto, resolveErr := libraryResolver.GetLibraryByCategory(category)
		if resolveErr != nil {
			return ExtractResult{}, fmt.Errorf("resolve library %s: %w", category, resolveErr)
		}

		destinationPath, pathErr := buildDestinationPath(libraryDto.Path, entryName)
		if pathErr != nil {
			applog.Warn("takeout: skipping entry with unsafe destination", "entry", entry.Name)
			result.SkippedEntries++
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
			return ExtractResult{}, fmt.Errorf("create destination directory: %w", err)
		}

		content, readErr := readZipEntry(entry)
		if readErr != nil {
			return ExtractResult{}, readErr
		}
		if writeErr := os.WriteFile(destinationPath, content, 0644); writeErr != nil {
			return ExtractResult{}, fmt.Errorf("write extracted file: %w", writeErr)
		}

		metadata, ok := metadataByName[entryName]
		var metadataPointer *TakeoutMetadata
		if ok {
			metadataCopy := metadata
			metadataPointer = &metadataCopy
		}

		result.Files = append(result.Files, ExtractedFile{
			SourcePath:      entry.Name,
			DestinationPath: destinationPath,
			Category:        string(category),
			Metadata:        metadataPointer,
		})
	}

	return result, nil
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
