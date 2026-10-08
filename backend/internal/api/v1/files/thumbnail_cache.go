package files

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultThumbnailWidth = 320
	maxThumbnailWidth     = 2048
)

func normalizeThumbnailWidth(width int) int {
	if width <= 0 {
		return defaultThumbnailWidth
	}
	return min(width, maxThumbnailWidth)
}

func thumbnailCacheFilePrefix(fileDto FileDto, width int) string {
	return fmt.Sprintf("%d_%d_", fileDto.ID, normalizeThumbnailWidth(width))
}

func thumbnailCacheFileName(fileDto FileDto, width int) string {
	return fmt.Sprintf("%s%d.png", thumbnailCacheFilePrefix(fileDto, width), fileDto.UpdatedAt.UnixNano())
}

// ThumbnailETag identifies one rendition of a file's thumbnail. It changes
// whenever the file's updated_at changes, so a modified file never revalidates
// against a stale thumbnail.
func ThumbnailETag(fileDto FileDto, width int) string {
	return fmt.Sprintf(`"%s"`, thumbnailCacheFileName(fileDto, width))
}

func removeStaleThumbnails(cacheDir string, fileDto FileDto, width int) {
	staleCandidates, err := filepath.Glob(filepath.Join(cacheDir, thumbnailCacheFilePrefix(fileDto, width)+"*.png"))
	if err != nil {
		return
	}
	currentName := thumbnailCacheFileName(fileDto, width)
	for _, candidate := range staleCandidates {
		if filepath.Base(candidate) != currentName {
			_ = os.Remove(candidate)
		}
	}
}
