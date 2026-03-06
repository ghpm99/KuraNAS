package files

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/config"

	"github.com/gin-gonic/gin"
)

func TestResolvePathInEntryPoint(t *testing.T) {
	tempDir := t.TempDir()
	config.AppConfig.EntryPoint = tempDir

	path, err := resolvePathInEntryPoint("")
	if err != nil {
		t.Fatalf("expected entry point path, got error: %v", err)
	}
	if path != filepath.Clean(tempDir) {
		t.Fatalf("expected %s, got %s", filepath.Clean(tempDir), path)
	}

	relativePath := "docs/file.txt"
	resolvedRelative, err := resolvePathInEntryPoint(relativePath)
	if err != nil {
		t.Fatalf("expected valid relative path, got error: %v", err)
	}
	expectedRelative := filepath.Join(filepath.Clean(tempDir), filepath.FromSlash(relativePath))
	if resolvedRelative != expectedRelative {
		t.Fatalf("expected %s, got %s", expectedRelative, resolvedRelative)
	}

	if _, err := resolvePathInEntryPoint("../outside"); err == nil {
		t.Fatalf("expected error for path outside entry point")
	}
}

func TestUploadFilesHandlerReturnsAcceptedWithJobID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tempDir := t.TempDir()
	config.AppConfig.EntryPoint = tempDir

	handler := NewHandler(&filesHandlerServiceMock{}, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/upload", handler.UploadFilesHandler)

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	if err := writer.WriteField("target_path", "."); err != nil {
		t.Fatalf("failed to write target_path: %v", err)
	}

	part, err := writer.CreateFormFile("files", "example.txt")
	if err != nil {
		t.Fatalf("failed to create multipart file field: %v", err)
	}
	if _, err := part.Write([]byte("hello world")); err != nil {
		t.Fatalf("failed to write multipart file payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/upload", requestBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", w.Code, w.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response payload: %v", err)
	}
	jobID, ok := payload["job_id"].(string)
	if !ok || jobID == "" {
		t.Fatalf("expected non-empty job_id in response, payload=%v", payload)
	}

	uploadedPath := filepath.Join(tempDir, "example.txt")
	if _, err := os.Stat(uploadedPath); err != nil {
		t.Fatalf("expected uploaded file to be saved on disk: %v", err)
	}
}
