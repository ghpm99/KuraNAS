package captures

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Service mock — success
// ---------------------------------------------------------------------------

type capturesHandlerServiceMock struct{}

func (m *capturesHandlerServiceMock) UploadCapture(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error) {
	return CaptureDto{ID: 1, Name: dto.Name, FileName: file.Filename}, nil
}

func (m *capturesHandlerServiceMock) InitCaptureUpload(dto InitCaptureUploadDto) (InitCaptureUploadResultDto, error) {
	return InitCaptureUploadResultDto{UploadID: "abc", ChunkSize: 1024}, nil
}

func (m *capturesHandlerServiceMock) UploadCaptureChunk(file *multipart.FileHeader, dto UploadCaptureChunkDto) error {
	return nil
}

func (m *capturesHandlerServiceMock) CompleteCaptureUpload(dto CompleteCaptureUploadDto) (CaptureDto, error) {
	return CaptureDto{ID: 1, Name: "done", FileName: "file.bin"}, nil
}

func (m *capturesHandlerServiceMock) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error) {
	return utils.PaginationResponse[CaptureDto]{Items: []CaptureDto{{ID: 1, Name: "test"}}}, nil
}

func (m *capturesHandlerServiceMock) GetCaptureByID(id int) (CaptureDto, error) {
	return CaptureDto{ID: id, Name: "test"}, nil
}

func (m *capturesHandlerServiceMock) DeleteCapture(id int) error {
	return nil
}

// ---------------------------------------------------------------------------
// Service mock — error
// ---------------------------------------------------------------------------

type capturesHandlerErrServiceMock struct {
	capturesHandlerServiceMock
}

func (m *capturesHandlerErrServiceMock) UploadCapture(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error) {
	return CaptureDto{}, errors.New("upload failed")
}

func (m *capturesHandlerErrServiceMock) InitCaptureUpload(dto InitCaptureUploadDto) (InitCaptureUploadResultDto, error) {
	return InitCaptureUploadResultDto{}, errors.New("init failed")
}

func (m *capturesHandlerErrServiceMock) UploadCaptureChunk(file *multipart.FileHeader, dto UploadCaptureChunkDto) error {
	return errors.New("chunk failed")
}

func (m *capturesHandlerErrServiceMock) CompleteCaptureUpload(dto CompleteCaptureUploadDto) (CaptureDto, error) {
	return CaptureDto{}, errors.New("complete failed")
}

func (m *capturesHandlerErrServiceMock) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error) {
	return utils.PaginationResponse[CaptureDto]{}, errors.New("list failed")
}

func (m *capturesHandlerErrServiceMock) GetCaptureByID(id int) (CaptureDto, error) {
	return CaptureDto{}, errors.New("not found")
}

func (m *capturesHandlerErrServiceMock) DeleteCapture(id int) error {
	return errors.New("delete failed")
}

// ---------------------------------------------------------------------------
// Logger mock
// ---------------------------------------------------------------------------

type capturesLoggerMock struct{ logger.LoggerServiceInterface }

func (m *capturesLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *capturesLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (m *capturesLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helper: build multipart request
// ---------------------------------------------------------------------------

func buildMultipartUploadRequest(t *testing.T, path string, name string, includeFile bool) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if includeFile {
		part, err := writer.CreateFormFile("file", "video.mp4")
		if err != nil {
			t.Fatal(err)
		}
		part.Write([]byte("fake-video-data"))
	}

	writer.WriteField("name", name)
	writer.WriteField("media_type", "hls")
	writer.WriteField("mime_type", "video/mp2t")
	writer.WriteField("size", "15")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func buildJSONRequest(t *testing.T, method string, path string, payload any) *http.Request {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func buildChunkUploadRequest(t *testing.T, path string, uploadID string, offset int64, includeChunk bool) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("upload_id", uploadID)
	_ = writer.WriteField("offset", fmt.Sprintf("%d", offset))

	if includeChunk {
		part, err := writer.CreateFormFile("chunk", "chunk.bin")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write([]byte("chunk-data"))
	}

	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// ---------------------------------------------------------------------------
// Tests — Success paths
// ---------------------------------------------------------------------------

func TestCapturesHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&capturesHandlerServiceMock{}, &capturesLoggerMock{})
	router := gin.New()

	router.POST("/captures/upload", handler.UploadCaptureHandler)
	router.POST("/captures/upload/init", handler.InitCaptureUploadHandler)
	router.POST("/captures/upload/chunk", handler.UploadCaptureChunkHandler)
	router.POST("/captures/upload/complete", handler.CompleteCaptureUploadHandler)
	router.GET("/captures", handler.GetCapturesHandler)
	router.GET("/captures/:id", handler.GetCaptureByIDHandler)
	router.DELETE("/captures/:id", handler.DeleteCaptureHandler)

	t.Run("UploadCapture success", func(t *testing.T) {
		req := buildMultipartUploadRequest(t, "/captures/upload", "my_video", true)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, w.Code, w.Body.String())
		}
	})

	t.Run("GetCaptures success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/captures?page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("InitCaptureUpload success", func(t *testing.T) {
		req := buildJSONRequest(t, http.MethodPost, "/captures/upload/init", InitCaptureUploadDto{
			Name: "my_video",
			Size: 100,
		})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("UploadCaptureChunk success", func(t *testing.T) {
		req := buildChunkUploadRequest(t, "/captures/upload/chunk", "abc", 0, true)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("CompleteCaptureUpload success", func(t *testing.T) {
		req := buildJSONRequest(t, http.MethodPost, "/captures/upload/complete", CompleteCaptureUploadDto{
			UploadID: "abc",
		})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("GetCaptures with filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/captures?name=test&media_type=hls", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("GetCaptureByID success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/captures/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("DeleteCapture success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/captures/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Error paths
// ---------------------------------------------------------------------------

func TestCapturesHandlerErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&capturesHandlerErrServiceMock{}, &capturesLoggerMock{})
	router := gin.New()

	router.POST("/captures/upload", handler.UploadCaptureHandler)
	router.POST("/captures/upload/init", handler.InitCaptureUploadHandler)
	router.POST("/captures/upload/chunk", handler.UploadCaptureChunkHandler)
	router.POST("/captures/upload/complete", handler.CompleteCaptureUploadHandler)
	router.GET("/captures", handler.GetCapturesHandler)
	router.GET("/captures/:id", handler.GetCaptureByIDHandler)
	router.DELETE("/captures/:id", handler.DeleteCaptureHandler)

	t.Run("UploadCapture service error", func(t *testing.T) {
		req := buildMultipartUploadRequest(t, "/captures/upload", "my_video", true)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("UploadCapture no file", func(t *testing.T) {
		req := buildMultipartUploadRequest(t, "/captures/upload", "my_video", false)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("UploadCapture no name", func(t *testing.T) {
		req := buildMultipartUploadRequest(t, "/captures/upload", "", true)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("GetCaptures error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/captures", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("InitCaptureUpload error", func(t *testing.T) {
		req := buildJSONRequest(t, http.MethodPost, "/captures/upload/init", InitCaptureUploadDto{Name: "x"})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("UploadCaptureChunk no file", func(t *testing.T) {
		req := buildChunkUploadRequest(t, "/captures/upload/chunk", "abc", 0, false)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CompleteCaptureUpload error", func(t *testing.T) {
		req := buildJSONRequest(t, http.MethodPost, "/captures/upload/complete", CompleteCaptureUploadDto{
			UploadID: "abc",
		})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("GetCaptureByID not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/captures/99", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("GetCaptureByID invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/captures/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("DeleteCapture error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/captures/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("DeleteCapture invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/captures/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — NewHandler constructor
// ---------------------------------------------------------------------------

func TestNewHandlerReturnsNonNil(t *testing.T) {
	handler := NewHandler(&capturesHandlerServiceMock{}, &capturesLoggerMock{})
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ---------------------------------------------------------------------------
// Tests — UploadCapture with size parsing
// ---------------------------------------------------------------------------

func TestUploadCaptureSizeParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedDto CreateCaptureDto
	mock := &capturesSizeMock{
		uploadFn: func(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error) {
			capturedDto = dto
			return CaptureDto{ID: 1, Name: dto.Name}, nil
		},
	}

	handler := NewHandler(mock, &capturesLoggerMock{})
	router := gin.New()
	router.POST("/captures/upload", handler.UploadCaptureHandler)

	req := buildMultipartUploadRequest(t, "/captures/upload", "test", true)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
	}
	if capturedDto.Size != 15 {
		t.Fatalf("expected size 15, got %d", capturedDto.Size)
	}
	if capturedDto.MediaType != "hls" {
		t.Fatalf("expected media_type hls, got %s", capturedDto.MediaType)
	}
}

type capturesSizeMock struct {
	capturesHandlerServiceMock
	uploadFn func(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error)
}

func (m *capturesSizeMock) UploadCapture(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error) {
	if m.uploadFn != nil {
		return m.uploadFn(file, dto)
	}
	return CaptureDto{}, fmt.Errorf("not implemented")
}
