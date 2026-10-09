package files

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultThumbnailSize = 320
	maxThumbnailSize     = 2048

	GridThumbnailSize = 400

	thumbnailJPEGExtension = "jpg"
	thumbnailPNGExtension  = "png"
)

var thumbnailCacheExtensions = []string{thumbnailJPEGExtension, thumbnailPNGExtension}

func normalizeThumbnailSize(size int) int {
	if size <= 0 {
		return defaultThumbnailSize
	}
	return min(size, maxThumbnailSize)
}

func thumbnailCacheFilePrefix(fileDto FileDto, width, height int) string {
	return fmt.Sprintf("%d_%dx%d_", fileDto.ID, normalizeThumbnailSize(width), normalizeThumbnailSize(height))
}

func thumbnailCacheVersion(fileDto FileDto, width, height int) string {
	return fmt.Sprintf("%s%d", thumbnailCacheFilePrefix(fileDto, width, height), fileDto.UpdatedAt.UnixNano())
}

func thumbnailCacheFileName(fileDto FileDto, width, height int, extension string) string {
	return fmt.Sprintf("%s.%s", thumbnailCacheVersion(fileDto, width, height), extension)
}

// ThumbnailETag identifies one rendition (id, box size and file version) of a
// file's thumbnail. It changes whenever the file's updated_at changes, so a
// modified file never revalidates against a stale thumbnail.
func ThumbnailETag(fileDto FileDto, width, height int) string {
	return fmt.Sprintf(`"%s"`, thumbnailCacheVersion(fileDto, width, height))
}

func readCachedThumbnail(cacheDir string, fileDto FileDto, width, height int) ([]byte, bool) {
	for _, extension := range thumbnailCacheExtensions {
		cachedData, err := os.ReadFile(filepath.Join(cacheDir, thumbnailCacheFileName(fileDto, width, height, extension)))
		if err == nil {
			return cachedData, true
		}
	}
	return nil, false
}

func removeStaleThumbnails(cacheDir string, fileDto FileDto, width, height int, currentExtension string) {
	currentName := thumbnailCacheFileName(fileDto, width, height, currentExtension)
	currentPrefix := thumbnailCacheFilePrefix(fileDto, width, height)
	removeCacheFilesMatching(cacheDir, currentPrefix+"*", currentName)

	legacyPrefix := fmt.Sprintf("%d_%d_", fileDto.ID, normalizeThumbnailSize(width))
	removeCacheFilesMatching(cacheDir, legacyPrefix+"*.png", currentName)
}

func removeCacheFilesMatching(cacheDir, pattern, keptName string) {
	candidates, err := filepath.Glob(filepath.Join(cacheDir, pattern))
	if err != nil {
		return
	}
	for _, candidate := range candidates {
		if filepath.Base(candidate) != keptName {
			_ = os.Remove(candidate)
		}
	}
}
