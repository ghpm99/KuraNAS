package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/config"
)

func TestHasCachedThumbnailMatchesOnlyTheCurrentFileVersion(t *testing.T) {
	cacheDir := config.GetBuildConfig("ThumbnailPath")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	cachedFile := FileDto{ID: 987654321, UpdatedAt: time.Unix(1700000000, 0)}
	cachedPath := filepath.Join(cacheDir, thumbnailCacheFileName(cachedFile, GridThumbnailSize, GridThumbnailSize, thumbnailJPEGExtension))

	if HasCachedThumbnail(cachedFile) {
		t.Fatalf("expected no cached thumbnail before it is written")
	}
	if err := os.WriteFile(cachedPath, []byte("jpeg"), 0o644); err != nil {
		t.Fatalf("write cached thumbnail: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(cachedPath) })

	if !HasCachedThumbnail(cachedFile) {
		t.Fatalf("expected the written thumbnail to be found")
	}
	modifiedFile := cachedFile
	modifiedFile.UpdatedAt = cachedFile.UpdatedAt.Add(time.Second)
	if HasCachedThumbnail(modifiedFile) {
		t.Fatalf("a modified file must not match the stale thumbnail")
	}
}
