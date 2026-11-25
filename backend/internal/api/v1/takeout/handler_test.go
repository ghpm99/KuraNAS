package takeout

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type handlerServiceMock struct {
	initFn      func(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error)
	chunkFn     func(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error
	completeFn  func(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error)
	statusFn    func(uploadID string) (TakeoutUploadStatusDto, error)
	cancelFn    func(uploadID string) error
	listFn      func(page int, pageSize int) (utils.PaginationResponse[TakeoutImportDto], error)
	getImportFn func(importID int) (TakeoutImportDto, error)
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
	return TakeoutImportResultDto{JobID: 12, ImportID: 5, Message: "ok"}, nil
}

func (m *handlerServiceMock) GetUploadStatus(uploadID string) (TakeoutUploadStatusDto, error) {
	if m.statusFn != nil {
		return m.statusFn(uploadID)
	}
	return TakeoutUploadStatusDto{UploadID: uploadID, Size: 10, ReceivedSize: 4}, nil
}

func (m *handlerServiceMock) CancelUpload(uploadID string) error {
	if m.cancelFn != nil {
		return m.cancelFn(uploadID)
	}
	return nil
}

func (m *handlerServiceMock) ListImports(page int, pageSize int) (utils.PaginationResponse[TakeoutImportDto], error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize)
	}
	return utils.PaginationResponse[TakeoutImportDto]{}, nil
}

func (m *handlerServiceMock) GetImport(importID int) (TakeoutImportDto, error) {
	if m.getImportFn != nil {
		return m.getImportFn(importID)
	}
	return TakeoutImportDto{ID: importID}, nil
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
		var response struct {
			JobID    int `json:"job_id"`
			ImportID int `json:"import_id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if response.JobID != 12 || response.ImportID != 5 {
			t.Fatalf("unexpected response: %+v", response)
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

	// Pins the request seam across all three takeout bodies: init (JSON),
	// chunk (multipart form) and complete (JSON). A json/form tag drift fails
	// here instead of silently dropping a field during the upload handshake.
	t.Run("decodes init/chunk/complete payloads", func(t *testing.T) {
		var initDto InitTakeoutUploadDto
		var chunkDto UploadTakeoutChunkDto
		var completeDto CompleteTakeoutUploadDto
		handler := NewHandler(&handlerServiceMock{
			initFn: func(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
				initDto = dto
				return InitTakeoutUploadResultDto{UploadID: "u1", ChunkSize: 1024}, nil
			},
			chunkFn: func(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error {
				chunkDto = dto
				return nil
			},
			completeFn: func(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
				completeDto = dto
				return TakeoutImportResultDto{JobID: 12, ImportID: 5, Message: "ok"}, nil
			},
		}, nil)

		recInit := httptest.NewRecorder()
		ctxInit, _ := gin.CreateTestContext(recInit)
		ctxInit.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", `{"file_name":"takeout.zip","size":2048}`)
		handler.InitUploadHandler(ctxInit)
		if initDto.FileName != "takeout.zip" || initDto.Size != 2048 {
			t.Fatalf("init payload did not decode: %+v", initDto)
		}

		recChunk := httptest.NewRecorder()
		ctxChunk, _ := gin.CreateTestContext(recChunk)
		ctxChunk.Request = buildChunkRequest(t, "u1", "1024", true)
		handler.UploadChunkHandler(ctxChunk)
		if chunkDto.UploadID != "u1" || chunkDto.Offset != 1024 {
			t.Fatalf("chunk form did not decode: %+v", chunkDto)
		}

		recDone := httptest.NewRecorder()
		ctxDone, _ := gin.CreateTestContext(recDone)
		ctxDone.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u9"}`)
		handler.CompleteUploadHandler(ctxDone)
		if completeDto.UploadID != "u9" {
			t.Fatalf("complete payload did not decode: %+v", completeDto)
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

	t.Run("UploadChunkHandler chunk exceeds declared size", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = buildChunkRequest(t, "u1", "0", true)

		handler := NewHandler(&handlerServiceMock{
			chunkFn: func(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error {
				return ErrUploadChunkExceedsSize
			},
		}, &handlerLoggerMock{})
		handler.UploadChunkHandler(ctx)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		expectedMessage := i18n.GetMessage("TAKEOUT_UPLOAD_CHUNK_EXCEEDS_SIZE")
		if body["error"] != expectedMessage || expectedMessage == "" {
			t.Fatalf("unexpected error message %q", body["error"])
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

func TestInitUploadHandlerRejectsInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name         string
		body         string
		serviceError error
		messageKey   string
		expectedCode int
	}{
		{"size zero", `{"file_name":"takeout.zip","size":0}`, ErrInvalidUploadSize, "TAKEOUT_UPLOAD_SIZE_INVALID", http.StatusBadRequest},
		{"rar name", `{"file_name":"takeout.rar","size":10}`, ErrInvalidUploadFileName, "TAKEOUT_UPLOAD_FILE_NAME_INVALID", http.StatusBadRequest},
		{"insufficient storage", `{"file_name":"takeout.zip","size":10}`, ErrInsufficientStorage, "TAKEOUT_UPLOAD_INSUFFICIENT_STORAGE", http.StatusInsufficientStorage},
		{"uppercase zip", `{"file_name":"FOTOS.ZIP","size":10}`, nil, "", http.StatusOK},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", testCase.body)
			service := &handlerServiceMock{initFn: func(InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
				return InitTakeoutUploadResultDto{UploadID: "u1"}, testCase.serviceError
			}}

			NewHandler(service, nil).InitUploadHandler(ctx)

			if rec.Code != testCase.expectedCode {
				t.Fatalf("expected %d, got %d", testCase.expectedCode, rec.Code)
			}
			if testCase.messageKey == "" {
				return
			}
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["error"] != i18n.GetMessage(testCase.messageKey) || body["error"] == "" {
				t.Fatalf("unexpected error message %q", body["error"])
			}
		})
	}
}

func TestInitUploadHandlerReturnsResumeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", `{"file_name":"takeout.zip","size":10}`)
	service := &handlerServiceMock{initFn: func(InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
		return InitTakeoutUploadResultDto{UploadID: "u1", ChunkSize: 2048, ReceivedSize: 4, Resumed: true}, nil
	}}

	NewHandler(service, nil).InitUploadHandler(ctx)

	var body struct {
		UploadID     string `json:"upload_id"`
		ChunkSize    int64  `json:"chunk_size"`
		ReceivedSize int64  `json:"received_size"`
		Resumed      bool   `json:"resumed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.UploadID != "u1" || body.ChunkSize != 2048 || body.ReceivedSize != 4 || !body.Resumed {
		t.Fatalf("unexpected body %+v", body)
	}
}

func TestGetUploadStatusHandlerReturnsSessionProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/takeout/upload/:upload_id", NewHandler(&handlerServiceMock{}, nil).GetUploadStatusHandler)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/takeout/upload/abc", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		UploadID     string `json:"upload_id"`
		Size         int64  `json:"size"`
		ReceivedSize int64  `json:"received_size"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.UploadID != "abc" || body.Size != 10 || body.ReceivedSize != 4 {
		t.Fatalf("unexpected body %+v", body)
	}
}

func TestGetUploadStatusHandlerReturns404ForUnknownSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &handlerServiceMock{statusFn: func(string) (TakeoutUploadStatusDto, error) {
		return TakeoutUploadStatusDto{}, ErrUploadSessionNotFound
	}}
	router := gin.New()
	router.GET("/takeout/upload/:upload_id", NewHandler(service, nil).GetUploadStatusHandler)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/takeout/upload/abc", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error == "" {
		t.Fatalf("expected error message, got %q (%v)", rec.Body.String(), err)
	}
}

func TestCancelUploadHandlerReturns204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/takeout/upload/:upload_id", NewHandler(&handlerServiceMock{}, nil).CancelUploadHandler)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/takeout/upload/abc", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestCancelUploadHandlerReturns404ForUnknownSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &handlerServiceMock{cancelFn: func(string) error { return ErrUploadSessionNotFound }}
	router := gin.New()
	router.DELETE("/takeout/upload/:upload_id", NewHandler(service, nil).CancelUploadHandler)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/takeout/upload/abc", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCancelUploadHandlerReturns500WithoutLeakingRawError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &handlerServiceMock{cancelFn: func(string) error { return errors.New("disk exploded") }}
	router := gin.New()
	router.DELETE("/takeout/upload/:upload_id", NewHandler(service, nil).CancelUploadHandler)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/takeout/upload/abc", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "disk exploded") {
		t.Fatalf("raw error leaked: %s", rec.Body.String())
	}
}
