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

func extractEntryWithoutOverwrite(entry *zip.File, destinationPath string) (string, bool, error) {
	content, err := readZipEntry(entry)
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return "", false, fmt.Errorf("create destination directory: %w", err)
	}

	contentHash := sha256.Sum256(content)
	for attempt := 1; ; attempt++ {
		candidatePath := suffixedPath(destinationPath, attempt)
		writeErr := writeNewFile(candidatePath, content)
		if writeErr == nil {
			return candidatePath, false, nil
		}
		if !errors.Is(writeErr, os.ErrExist) {
			return "", false, fmt.Errorf("write extracted file: %w", writeErr)
		}
		isSameContent, err := existingFileHasContent(candidatePath, int64(len(content)), contentHash)
		if err != nil {
			return "", false, err
		}
		if isSameContent {
			return candidatePath, true, nil
		}
	}
}

func suffixedPath(destinationPath string, attempt int) string {
	if attempt == 1 {
		return destinationPath
	}
	extension := filepath.Ext(destinationPath)
	baseName := strings.TrimSuffix(destinationPath, extension)
	return fmt.Sprintf("%s (%d)%s", baseName, attempt, extension)
}

func writeNewFile(filePath string, content []byte) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		return err
	}
	return file.Close()
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
