package takeout

import (
	"bytes"
	"errors"
	"mime/multipart"
	"nas-go/api/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type handlerServiceMock struct {
	initFn     func(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error)
	chunkFn    func(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error
	completeFn func(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error)
}

type handlerLoggerMock struct{}

func (m *handlerLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return log, nil
}
func (m *handlerLoggerMock) GetLogByID(id int) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *handlerLoggerMock) GetLogs(page, pageSize int) ([]logger.LoggerModel, error) {
	return nil, nil
}
func (m *handlerLoggerMock) UpdateLog(log logger.LoggerModel) error { return nil }
func (m *handlerLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error {
	return nil
}
func (m *handlerLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func (m *handlerServiceMock) InitUpload(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
	if m.initFn != nil {
		return m.initFn(dto)
	}
	return InitTakeoutUploadResultDto{UploadID: "u1", ChunkSize: 1024}, nil
}

func (m *handlerServiceMock) UploadChunk(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error {
	if m.chunkFn != nil {
		return m.chunkFn(file, dto)
	}
	return nil
}

func (m *handlerServiceMock) CompleteUpload(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
	if m.completeFn != nil {
		return m.completeFn(dto)
	}
	return TakeoutImportResultDto{JobID: 12, Message: "ok"}, nil
}

func buildJSONRequest(method string, url string, body string) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func buildChunkRequest(t *testing.T, uploadID string, offset string, includeChunk bool) *http.Request {
	t.Helper()
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	if includeChunk {
		part, err := writer.CreateFormFile("chunk", "chunk.bin")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		_, _ = part.Write([]byte("hello"))
	}
	_ = writer.WriteField("upload_id", uploadID)
	_ = writer.WriteField("offset", offset)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/takeout/upload/chunk", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestTakeoutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("InitUploadHandler success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", `{"file_name":"takeout.zip","size":10}`)

		handler := NewHandler(&handlerServiceMock{}, nil)
		handler.InitUploadHandler(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("InitUploadHandler bad request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", `{`)

		handler := NewHandler(&handlerServiceMock{}, nil)
		handler.InitUploadHandler(ctx)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("UploadChunkHandler success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildChunkRequest(t, "u1", "0", true)

		handler := NewHandler(&handlerServiceMock{}, nil)
		handler.UploadChunkHandler(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("UploadChunkHandler no file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildChunkRequest(t, "u1", "0", false)

		handler := NewHandler(&handlerServiceMock{}, nil)
		handler.UploadChunkHandler(ctx)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("CompleteUploadHandler success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u1"}`)

		handler := NewHandler(&handlerServiceMock{}, nil)
		handler.CompleteUploadHandler(ctx)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d", rec.Code)
		}
	})

	t.Run("CompleteUploadHandler bad request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{`)

		handler := NewHandler(&handlerServiceMock{}, nil)
		handler.CompleteUploadHandler(ctx)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("CompleteUploadHandler service error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u1"}`)

		handler := NewHandler(&handlerServiceMock{
			completeFn: func(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
				return TakeoutImportResultDto{}, errors.New("boom")
			},
		}, nil)
		handler.CompleteUploadHandler(ctx)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("UploadChunkHandler offset mismatch", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildChunkRequest(t, "u1", "0", true)

		handler := NewHandler(&handlerServiceMock{
			chunkFn: func(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error {
				return ErrUploadOffsetMismatch
			},
		}, &handlerLoggerMock{})
		handler.UploadChunkHandler(ctx)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("CompleteUploadHandler invalid zip", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u1"}`)

		handler := NewHandler(&handlerServiceMock{
			completeFn: func(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
				return TakeoutImportResultDto{}, ErrInvalidZipFile
			},
		}, &handlerLoggerMock{})
		handler.CompleteUploadHandler(ctx)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}
