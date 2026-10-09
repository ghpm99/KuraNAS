package files

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"nas-go/api/internal/roots"
)

func hasTraversalSegment(path string) bool {
	for _, segment := range strings.Split(normalizePathSeparators(path), "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

func isInsideDirectory(directoryPath string, candidatePath string) bool {
	relativePath, err := filepath.Rel(directoryPath, candidatePath)
	if err != nil {
		return false
	}
	return relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}

func normalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func resolveContainedPath(clientPath string) (string, error) {
	if strings.ContainsRune(clientPath, 0) || hasTraversalSegment(clientPath) {
		return "", fmt.Errorf("path %q contains a traversal sequence", clientPath)
	}

	absolutePath, err := roots.ResolveAbsolute(normalizePathSeparators(strings.TrimSpace(clientPath)))
	if err != nil {
		return "", err
	}

	cleanPath := filepath.Clean(absolutePath)
	owner, isOwned := roots.OwnerOf(cleanPath)
	if !isOwned || !isInsideDirectory(owner.Path, cleanPath) {
		return "", fmt.Errorf("path %q is outside every storage root", clientPath)
	}
	return cleanPath, nil
}

func resolveContainedChildPath(parentAbsolutePath string, childName string) (string, error) {
	if childName == "" || childName == "." || childName == ".." || childName != filepath.Base(normalizePathSeparators(childName)) || strings.ContainsAny(childName, "/\\") {
		return "", fmt.Errorf("invalid child name %q", childName)
	}
	return resolveContainedPath(filepath.Join(parentAbsolutePath, childName))
}

func resolveColdContentPath(coldPath string) (string, error) {
	if strings.ContainsRune(coldPath, 0) || hasTraversalSegment(coldPath) || !filepath.IsAbs(coldPath) {
		return "", fmt.Errorf("invalid cold path %q", coldPath)
	}
	cleanPath := filepath.Clean(coldPath)
	if _, isInsideRoot := roots.OwnerOf(cleanPath); isInsideRoot {
		return "", fmt.Errorf("cold path %q is inside a storage root", coldPath)
	}
	return cleanPath, nil
}

func invalidPathOperationError(err error) *FileOperationError {
	return newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
}

func resolveContainedSourcePath(sourcePath string) (string, error) {
	coldPath, err := resolveColdContentPath(sourcePath)
	if err == nil {
		return coldPath, nil
	}
	return resolveContainedPath(sourcePath)
}
