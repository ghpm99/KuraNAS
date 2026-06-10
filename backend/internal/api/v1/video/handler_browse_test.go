package video

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	files "nas-go/api/internal/api/v1/files"

	"github.com/gin-gonic/gin"
)

type videoFilesServiceMock struct {
	files.ServiceInterface
	filePath string
	format   string
	missing  bool
}

func (m *videoFilesServiceMock) GetFileById(id int) (files.FileDto, error) {
	if m.missing {
		return files.FileDto{}, errors.New("not found")
	}
	path := m.filePath
	if path == "" {
		path = "/tmp/missing.mp4"
	}
	format := m.format
	if format == "" {
		format = ".mp4"
	}
	return files.FileDto{
		ID:     id,
		Name:   "video",
		Path:   path,
		Format: format,
		Type:   files.File,
	}, nil
}

func (m *videoFilesServiceMock) CheckFileExistsByPath(path string) bool {
	return m.filePath != "" && path == m.filePath
}

func TestVideoHandlerBrowseEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&videoHandlerServiceMock{}, &videoFilesServiceMock{}, &videoLoggerMock{})

	router := gin.New()
	router.GET("/files/videos", handler.GetVideosHandler)
	router.GET("/files/video-thumbnail/:id", handler.GetVideoThumbnailHandler)
	router.GET("/files/video-preview/:id", handler.GetVideoPreviewHandler)
	router.GET("/files/video-stream/:id", handler.StreamVideoHandler)

	tests := []struct {
		path string
		code int
	}{
		{"/files/videos", http.StatusOK},
		{"/files/video-thumbnail/1", http.StatusOK},
		{"/files/video-preview/1", http.StatusOK},
		{"/files/video-stream/1", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestVideoHandlerBrowseErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&videoHandlerErrServiceMock{}, &videoFilesServiceMock{}, &videoLoggerMock{})

	router := gin.New()
	router.GET("/files/videos", handler.GetVideosHandler)
	router.GET("/files/video-thumbnail/:id", handler.GetVideoThumbnailHandler)
	router.GET("/files/video-preview/:id", handler.GetVideoPreviewHandler)

	tests := []struct {
		path string
		code int
	}{
		{"/files/videos", http.StatusInternalServerError},
		{"/files/video-thumbnail/1", http.StatusNotFound},
		{"/files/video-preview/1", http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}

	missingHandler := NewHandler(&videoHandlerErrServiceMock{}, &videoFilesServiceMock{missing: true}, &videoLoggerMock{})
	missingRouter := gin.New()
	missingRouter.GET("/files/video-thumbnail/:id", missingHandler.GetVideoThumbnailHandler)
	missingRouter.GET("/files/video-stream/:id", missingHandler.StreamVideoHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/video-thumbnail/1", nil)
	w := httptest.NewRecorder()
	missingRouter.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing file record, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	w = httptest.NewRecorder()
	missingRouter.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing stream record, got %d", w.Code)
	}
}

func TestVideoHandlerStreamRanges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "v.mp4")
	if err := os.WriteFile(videoPath, []byte("0123456789abcdefghijklmnopqrstuvwxyz"), 0644); err != nil {
		t.Fatalf("failed to create video file: %v", err)
	}

	filesService := &videoFilesServiceMock{filePath: videoPath, format: ".mp4"}
	handler := NewHandler(&videoHandlerServiceMock{}, filesService, &videoLoggerMock{})

	router := gin.New()
	router.GET("/files/video-stream/:id", handler.StreamVideoHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	req.Header.Set("Range", "bytes=0-10")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected partial content for video, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 0-10/36" {
		t.Fatalf("unexpected content-range for closed range: %s", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	req.Header.Set("Range", "bytes=0-")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected partial content for open-ended range, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 0-35/36" {
		t.Fatalf("unexpected content-range for open-ended range: %s", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	req.Header.Set("Range", "bytes=-5")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected partial content for suffix range, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 31-35/36" {
		t.Fatalf("unexpected content-range for suffix range: %s", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected full stream 200 without range, got %d", w.Code)
	}
}
