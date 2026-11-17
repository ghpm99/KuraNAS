package takeout

import (
	"archive/zip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var googlePhotosFolderNames = []string{"google fotos", "google photos"}

func stripGooglePhotosPrefix(entryName string) string {
	segments := splitEntrySegments(entryName)
	if len(segments) > 2 && strings.EqualFold(segments[0], "takeout") && isGooglePhotosFolder(segments[1]) {
		segments = segments[2:]
	}
	return strings.Join(segments, "/")
}

func isGooglePhotosFolder(folderName string) bool {
	for _, googlePhotosFolderName := range googlePhotosFolderNames {
		if strings.EqualFold(folderName, googlePhotosFolderName) {
			return true
		}
	}
	return false
}

func splitEntrySegments(entryPath string) []string {
	slashedPath := strings.ReplaceAll(entryPath, "\\", "/")
	segments := make([]string, 0)
	for _, segment := range strings.Split(slashedPath, "/") {
		if segment == "" || segment == "." {
			continue
		}
		segments = append(segments, segment)
	}
	return segments
}

func sanitizeEntryPath(entryPath string) string {
	segments := splitEntrySegments(entryPath)
	for index, segment := range segments {
		segments[index] = sanitizeTakeoutFileName(segment)
	}
	return strings.Join(segments, "/")
}

type stagedEntry struct {
	path string
	size int64
	hash [sha256.Size]byte
}

func extractEntryWithoutOverwrite(entry *zip.File, destinationPath string) (string, bool, error) {
	destinationDirectory := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destinationDirectory, 0755); err != nil {
		return "", false, fmt.Errorf("create destination directory: %w", err)
	}

	staged, err := stageEntryContent(entry, destinationDirectory)
	if err != nil {
		return "", false, err
	}
	defer os.Remove(staged.path)

	for attempt := 1; ; attempt++ {
		candidatePath := suffixedPath(destinationPath, attempt)
		isOccupied, err := isPathOccupied(candidatePath)
		if err != nil {
			return "", false, err
		}
		if !isOccupied {
			if err := os.Rename(staged.path, candidatePath); err != nil {
				return "", false, fmt.Errorf("move extracted file into place: %w", err)
			}
			return candidatePath, false, nil
		}
		isSameContent, err := existingFileHasContent(candidatePath, staged.size, staged.hash)
		if err != nil {
			return "", false, err
		}
		if isSameContent {
			return candidatePath, true, nil
		}
	}
}

func stageEntryContent(entry *zip.File, directory string) (stagedEntry, error) {
	source, err := entry.Open()
	if err != nil {
		return stagedEntry{}, fmt.Errorf("open zip entry: %w", err)
	}
	defer source.Close()

	partFile, err := os.CreateTemp(directory, ".takeout-*.part")
	if err != nil {
		return stagedEntry{}, fmt.Errorf("create partial file: %w", err)
	}

	hasher := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(partFile, hasher), source)
	closeErr := partFile.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		os.Remove(partFile.Name())
		return stagedEntry{}, fmt.Errorf("copy zip entry: %w", err)
	}
	if err := os.Chmod(partFile.Name(), 0644); err != nil {
		os.Remove(partFile.Name())
		return stagedEntry{}, fmt.Errorf("set partial file permissions: %w", err)
	}

	staged := stagedEntry{path: partFile.Name(), size: size}
	copy(staged.hash[:], hasher.Sum(nil))
	return staged, nil
}

func isPathOccupied(filePath string) (bool, error) {
	_, err := os.Lstat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func suffixedPath(destinationPath string, attempt int) string {
	if attempt == 1 {
		return destinationPath
	}
	extension := filepath.Ext(destinationPath)
	baseName := strings.TrimSuffix(destinationPath, extension)
	return fmt.Sprintf("%s (%d)%s", baseName, attempt, extension)
}

func existingFileHasContent(filePath string, expectedSize int64, expectedHash [sha256.Size]byte) (bool, error) {
	info, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Size() != expectedSize {
		return false, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, err
	}
	var existingHash [sha256.Size]byte
	copy(existingHash[:], hasher.Sum(nil))
	return existingHash == expectedHash, nil
}
