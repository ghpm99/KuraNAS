package video

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

func setProgramFilesForVideoTest(t *testing.T) string {
	t.Helper()

	programFiles := filepath.Join(t.TempDir(), "ProgramFiles")
	t.Setenv("ProgramFiles", programFiles)
	return programFiles
}

func ensureVideoTestIcons(t *testing.T) {
	t.Helper()

	setProgramFilesForVideoTest(t)

	iconDir := config.GetBuildConfig("IconPath")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		t.Fatalf("failed to create icon dir: %v", err)
	}

	writeIcon := func(name string) {
		t.Helper()
		path := filepath.Join(iconDir, name+".png")
		if _, err := os.Stat(path); err == nil {
			return
		}

		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("failed to create icon file %s: %v", path, err)
		}
		defer f.Close()

		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{R: 255, A: 255})
		img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("failed to encode icon %s: %v", path, err)
		}
	}

	for _, name := range []string{"folder", "unknown", "mp4", "mp3", "pdf"} {
		writeIcon(name)
	}
}

func newVideoBrowseServiceForTest(t *testing.T, repo *videoRepoMock) *Service {
	t.Helper()
	repo.db = database.NewDbContext(nil)
	return &Service{Repository: repo}
}

func TestVideoService_GetVideos(t *testing.T) {
	now := time.Now()
	s := newVideoBrowseServiceForTest(t, &videoRepoMock{
		getVideosFn: func(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
			return utils.PaginationResponse[files.FileModel]{Items: []files.FileModel{{
				ID: 3, Name: "v.mp4", Path: "/v/v.mp4", ParentPath: "/v",
				Format: ".mp4", Type: files.File, CreatedAt: now, UpdatedAt: now,
			}}}, nil
		},
	})

	out, err := s.GetVideos(1, 10)
	if err != nil || len(out.Items) != 1 {
		t.Fatalf("GetVideos failed len=%d err=%v", len(out.Items), err)
	}
}

func TestVideoService_ThumbnailAndPreviewFallbacks(t *testing.T) {
	ensureVideoTestIcons(t)

	tmpDir := t.TempDir()
	fakeVideo := filepath.Join(tmpDir, "video.mp4")
	if err := os.WriteFile(fakeVideo, []byte("not-a-real-video"), 0644); err != nil {
		t.Fatalf("failed to create fake video file: %v", err)
	}

	s := newVideoBrowseServiceForTest(t, &videoRepoMock{})

	videoThumb, err := s.GetVideoThumbnail(files.FileDto{
		ID:   103,
		Path: fakeVideo,
		Type: files.File,
	}, -1, -1)
	if err != nil {
		t.Fatalf("expected video thumbnail fallback success, got %v", err)
	}
	if len(videoThumb) == 0 {
		t.Fatalf("expected non-empty video thumbnail fallback")
	}

	previewGif, err := s.GetVideoPreviewGif(files.FileDto{
		ID:   104,
		Path: fakeVideo,
		Type: files.File,
	}, -1, -1)
	if err != nil {
		t.Fatalf("expected video preview fallback success, got %v", err)
	}
	if len(previewGif) == 0 {
		t.Fatalf("expected non-empty video preview fallback")
	}

	if _, err := s.GetVideoThumbnail(files.FileDto{
		ID:   105,
		Path: filepath.Join(tmpDir, "missing.mp4"),
		Type: files.File,
	}, 320, 180); err == nil {
		t.Fatalf("expected missing video thumbnail error")
	}

	if _, err := s.GetVideoPreviewGif(files.FileDto{
		ID:   106,
		Path: filepath.Join(tmpDir, "missing.mp4"),
		Type: files.File,
	}, 320, 180); err == nil {
		t.Fatalf("expected missing video preview error")
	}
}

func TestVideoService_ThumbAndPreviewCacheHit(t *testing.T) {
	setProgramFilesForVideoTest(t)
	s := newVideoBrowseServiceForTest(t, &videoRepoMock{})
	cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create video cache dir: %v", err)
	}

	thumbPath := filepath.Join(cacheDir, "501_320x180.png")
	thumbBytes := []byte("cached-video-thumb")
	if err := os.WriteFile(thumbPath, thumbBytes, 0644); err != nil {
		t.Fatalf("failed to write thumb cache: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(thumbPath) })

	thumb, err := s.GetVideoThumbnail(files.FileDto{ID: 501, Path: "/missing.mp4", Type: files.File}, 320, 180)
	if err != nil {
		t.Fatalf("expected video thumbnail cache hit, got %v", err)
	}
	if string(thumb) != string(thumbBytes) {
		t.Fatalf("expected cached thumbnail bytes")
	}

	previewPath := filepath.Join(cacheDir, "502_320x180_preview.gif")
	previewBytes := []byte("cached-video-preview")
	if err := os.WriteFile(previewPath, previewBytes, 0644); err != nil {
		t.Fatalf("failed to write preview cache: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(previewPath) })

	preview, err := s.GetVideoPreviewGif(files.FileDto{ID: 502, Path: "/missing.mp4", Type: files.File}, 320, 180)
	if err != nil {
		t.Fatalf("expected video preview cache hit, got %v", err)
	}
	if string(preview) != string(previewBytes) {
		t.Fatalf("expected cached preview bytes")
	}
}

func TestVideoService_ThumbnailUsesSourcePoster(t *testing.T) {
	ensureVideoTestIcons(t)
	s := newVideoBrowseServiceForTest(t, &videoRepoMock{})

	const fileID = 70123
	sourceDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video", "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create poster source dir: %v", err)
	}
	sourcePath := filepath.Join(sourceDir, fmt.Sprintf("%d", fileID))
	f, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("failed to create source poster: %v", err)
	}
	posterImg := image.NewRGBA(image.Rect(0, 0, 8, 8))
	posterImg.Set(0, 0, color.RGBA{G: 255, A: 255})
	if encErr := png.Encode(f, posterImg); encErr != nil {
		f.Close()
		t.Fatalf("failed to encode source poster: %v", encErr)
	}
	f.Close()
	t.Cleanup(func() { _ = os.Remove(sourcePath) })

	cachePath := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video", fmt.Sprintf("%d_320x180.png", fileID))
	t.Cleanup(func() { _ = os.Remove(cachePath) })

	// The video file itself is missing on disk; the source poster must still
	// yield a thumbnail (never reaching the missing-file error path).
	thumb, err := s.GetVideoThumbnail(files.FileDto{ID: fileID, Path: "/missing.mp4", Type: files.File}, 320, 180)
	if err != nil {
		t.Fatalf("expected source poster thumbnail, got %v", err)
	}
	if len(thumb) == 0 {
		t.Fatal("expected non-empty thumbnail from source poster")
	}
	if _, statErr := os.Stat(cachePath); statErr != nil {
		t.Fatalf("expected poster thumbnail to be cached: %v", statErr)
	}
}
