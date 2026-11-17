package files

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func (m *filesHandlerServiceMock) GetUploadSessionStatus(uploadID string) (UploadSessionStatus, error) {
	return UploadSessionStatus{}, nil
}

func newStatusRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/upload/init", handler.InitUploadHandler)
	router.POST("/files/upload/chunk", handler.UploadChunkHandler)
	router.GET("/files/upload/:upload_id", handler.GetUploadStatusHandler)
	return router
}

func getUploadStatus(router *gin.Engine, uploadID string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/files/upload/"+uploadID, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestGetUploadStatusReportsProgress(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newStatusRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 10)
	postChunk(t, router, uploadID, "0", "hello")

	recorder := getUploadStatus(router, uploadID)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		UploadID     string `json:"upload_id"`
		Size         int64  `json:"size"`
		ReceivedSize int64  `json:"received_size"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if response.UploadID != uploadID || response.Size != 10 || response.ReceivedSize != 5 {
		t.Fatalf("unexpected status: %+v", response)
	}
}

func TestGetUploadStatusUnknownSessionIsNotFound(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newStatusRouter(environment.service)

	recorder := getUploadStatus(router, strings.Repeat("a", 32))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetUploadStatusMalformedIDIsBadRequest(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newStatusRouter(environment.service)

	recorder := getUploadStatus(router, "not-an-id")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestInterruptedUploadResumesFromReceivedSize(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newStatusRouter(environment.service)
	content := "0123456789"
	body := `{"file_name":"a.bin","size":10,"checksum":"` + sha256HexOf(content) + `"}`
	firstInit := decodeInitResponse(t, postInitUpload(router, body))
	postChunk(t, router, firstInit.UploadID, "0", content[:4])

	resumedInit := decodeInitResponse(t, postInitUpload(router, body))
	postChunk(t, router, resumedInit.UploadID, "4", content[4:])

	if !resumedInit.Resumed || resumedInit.UploadID != firstInit.UploadID || resumedInit.ReceivedSize != 4 {
		t.Fatalf("unexpected resume: %+v", resumedInit)
	}
	if sha256HexOf(readPartialFile(t, firstInit.UploadID)) != sha256HexOf(content) {
		t.Fatalf("resumed content differs: %q", readPartialFile(t, firstInit.UploadID))
	}
}
