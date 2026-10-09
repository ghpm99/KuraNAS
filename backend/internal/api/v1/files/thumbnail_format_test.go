package files

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/config"

	"github.com/gin-gonic/gin"
)

func writeSourceImage(t *testing.T, path string, width, height int, alpha uint8, encodeAsJPEG bool) {
	t.Helper()
	canvas := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			canvas.Set(x, y, color.NRGBA{R: 200, G: 80, B: 40, A: alpha})
		}
	}
	var encoded bytes.Buffer
	if encodeAsJPEG {
		if err := jpeg.Encode(&encoded, canvas, nil); err != nil {
			t.Fatal(err)
		}
	} else if err := png.Encode(&encoded, canvas); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestThumbnailCacheKeyIncludesWidthAndHeight(t *testing.T) {
	fileDto := FileDto{ID: 5, UpdatedAt: time.Unix(10, 0)}

	square := thumbnailCacheFileName(fileDto, 960, 960, thumbnailJPEGExtension)
	landscape := thumbnailCacheFileName(fileDto, 960, 720, thumbnailJPEGExtension)

	if square == landscape {
		t.Fatalf("960x960 and 960x720 must not share a cache key")
	}
	if ThumbnailETag(fileDto, 960, 960) == ThumbnailETag(fileDto, 960, 720) {
		t.Fatalf("ETag must differ when only height differs")
	}
	if filepath.Ext(thumbnailCacheFileName(fileDto, 1, 1, thumbnailPNGExtension)) != ".png" {
		t.Fatalf("expected the cache name to carry the output format")
	}
}

func TestFileService_PhotoThumbnailIsJPEGFittedInsideBox(t *testing.T) {
	setProgramFilesForTest(t)
	photoPath := filepath.Join(t.TempDir(), "photo.jpg")
	writeSourceImage(t, photoPath, 800, 400, 255, true)
	s := newFilesServiceForTest(t, &filesRepoMock{})
	fileDto := FileDto{ID: 8101, Path: photoPath, Type: File, Format: ".jpg", UpdatedAt: time.Unix(1, 0)}

	data, err := s.GetFileThumbnail(fileDto, 200, 200)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("expected jpeg thumbnail: %v", err)
	}
	if decoded.Bounds().Dx() != 200 || decoded.Bounds().Dy() != 100 {
		t.Fatalf("expected 200x100 without letterbox, got %v", decoded.Bounds())
	}
	cacheDir := config.GetBuildConfig("ThumbnailPath")
	if _, err := os.Stat(filepath.Join(cacheDir, thumbnailCacheFileName(fileDto, 200, 200, thumbnailJPEGExtension))); err != nil {
		t.Fatalf("expected jpeg cache file: %v", err)
	}
	cachedAgain, err := s.GetFileThumbnail(fileDto, 200, 200)
	if err != nil || !bytes.Equal(cachedAgain, data) {
		t.Fatalf("expected second call to hit the cache")
	}
}

func TestFileService_TranslucentPNGThumbnailStaysPNG(t *testing.T) {
	setProgramFilesForTest(t)
	sourcePath := filepath.Join(t.TempDir(), "sticker.png")
	writeSourceImage(t, sourcePath, 100, 100, 100, false)
	s := newFilesServiceForTest(t, &filesRepoMock{})

	data, err := s.GetFileThumbnail(FileDto{ID: 8102, Path: sourcePath, Type: File, Format: ".png"}, 50, 50)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("expected png thumbnail for transparent source: %v", err)
	}
}

func TestFileService_RemovesLegacyThumbnailCacheFiles(t *testing.T) {
	setProgramFilesForTest(t)
	photoPath := filepath.Join(t.TempDir(), "photo.jpg")
	writeSourceImage(t, photoPath, 100, 100, 255, true)
	s := newFilesServiceForTest(t, &filesRepoMock{})
	cacheDir := config.GetBuildConfig("ThumbnailPath")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	removeCachedThumbnailsForFile(t, cacheDir, 8103)
	t.Cleanup(func() { removeCachedThumbnailsForFile(t, cacheDir, 8103) })
	legacyPath := filepath.Join(cacheDir, "8103_64_123.png")
	if err := os.WriteFile(legacyPath, []byte("legacy"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetFileThumbnail(FileDto{ID: 8103, Path: photoPath, Type: File, Format: ".jpg"}, 64, 64); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy cache file removed, got %v", err)
	}
}

func TestFilesHandlerThumbnailContentTypeMatchesEncodedFormat(t *testing.T) {
	var jpegBytes, pngBytes bytes.Buffer
	probe := image.NewRGBA(image.Rect(0, 0, 2, 2))
	_ = jpeg.Encode(&jpegBytes, probe, nil)
	_ = png.Encode(&pngBytes, probe)

	for expectedType, body := range map[string][]byte{"image/jpeg": jpegBytes.Bytes(), "image/png": pngBytes.Bytes()} {
		service := &filesHandlerServiceFuncMock{
			getFileByIdFn: func(id int) (FileDto, error) { return FileDto{ID: id, Type: File}, nil },
			getFileThumbnailFn: func(fileDto FileDto, width, height int) ([]byte, error) {
				return body, nil
			},
		}
		handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
		router := gin.New()
		router.GET("/files/thumbnail/:id", handler.GetFileThumbnailHandler)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/files/thumbnail/3?width=400&height=400", nil))

		if recorder.Header().Get("Content-Type") != expectedType {
			t.Fatalf("expected %s, got %s", expectedType, recorder.Header().Get("Content-Type"))
		}
	}
}

func removeCachedThumbnailsForFile(t *testing.T, cacheDir string, fileID int) {
	t.Helper()
	cachedPaths, err := filepath.Glob(filepath.Join(cacheDir, fmt.Sprintf("%d_*", fileID)))
	if err != nil {
		t.Fatal(err)
	}
	for _, cachedPath := range cachedPaths {
		_ = os.Remove(cachedPath)
	}
}
