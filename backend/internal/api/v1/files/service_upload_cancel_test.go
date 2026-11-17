package files

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func (m *filesHandlerServiceMock) CancelUpload(uploadID string) error {
	return nil
}

func newCancelRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/upload/init", handler.InitUploadHandler)
	router.POST("/files/upload/chunk", handler.UploadChunkHandler)
	router.DELETE("/files/upload/:upload_id", handler.CancelUploadHandler)
	return router
}

func deleteUpload(router *gin.Engine, uploadID string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodDelete, "/files/upload/"+uploadID, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCancelUploadRemovesSessionAndPartialFile(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCancelRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 10)
	postChunk(t, router, uploadID, "0", "hello")

	recorder := deleteUpload(router, uploadID)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if _, err := os.Stat(uploadSessionDir(uploadID)); !os.IsNotExist(err) {
		t.Fatalf("session dir must be gone, stat err = %v", err)
	}
	if _, err := os.Stat(uploadSessionPartialPath(uploadID)); !os.IsNotExist(err) {
		t.Fatalf("partial file must be gone, stat err = %v", err)
	}
}

func TestCancelUploadUnknownSessionIsNotFound(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCancelRouter(environment.service)

	recorder := deleteUpload(router, strings.Repeat("a", 32))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestCancelUploadMalformedIDIsBadRequest(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCancelRouter(environment.service)

	recorder := deleteUpload(router, "..x")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
