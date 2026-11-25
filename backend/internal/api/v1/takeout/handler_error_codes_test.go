package takeout

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func decodeErrorMessage(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body["error"]
}

func TestUploadChunkHandlerReturns404ForUnknownSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildChunkRequest(t, "unknown", "0", true)
	service := &handlerServiceMock{chunkFn: func(*multipart.FileHeader, UploadTakeoutChunkDto) error {
		return ErrUploadSessionNotFound
	}}

	NewHandler(service, nil).UploadChunkHandler(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
	if message := decodeErrorMessage(t, recorder); message != i18n.GetMessage("TAKEOUT_SESSION_NOT_FOUND") {
		t.Fatalf("unexpected message %q", message)
	}
}

func TestUploadChunkHandlerInternalErrorUsesTakeoutKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildChunkRequest(t, "u1", "0", true)
	service := &handlerServiceMock{chunkFn: func(*multipart.FileHeader, UploadTakeoutChunkDto) error {
		return errors.New("disk exploded")
	}}

	NewHandler(service, nil).UploadChunkHandler(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	message := decodeErrorMessage(t, recorder)
	if message != i18n.GetMessage("TAKEOUT_UPLOAD_CHUNK_ERROR") || strings.Contains(message, "disk exploded") {
		t.Fatalf("unexpected message %q", message)
	}
}

func TestCompleteUploadHandlerReturns404ForUnknownSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"unknown"}`)
	service := &handlerServiceMock{completeFn: func(CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
		return TakeoutImportResultDto{}, ErrUploadSessionNotFound
	}}

	NewHandler(service, nil).CompleteUploadHandler(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
	if message := decodeErrorMessage(t, recorder); message != i18n.GetMessage("TAKEOUT_SESSION_NOT_FOUND") {
		t.Fatalf("unexpected message %q", message)
	}
}

func TestCompleteUploadHandlerIncompleteMessageHasNoPlaceholders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.AppConfig.Lang = "pt-BR"
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("load translations: %v", err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u1"}`)
	service := &handlerServiceMock{completeFn: func(CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
		return TakeoutImportResultDto{}, UploadIncompleteError{ExpectedSize: 1000, ReceivedSize: 400}
	}}

	NewHandler(service, nil).CompleteUploadHandler(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	message := decodeErrorMessage(t, recorder)
	if strings.Contains(message, "%d") || !strings.Contains(message, "1000") || !strings.Contains(message, "400") {
		t.Fatalf("unexpected message %q", message)
	}
}

func TestInitUploadHandlerInternalErrorUsesTakeoutKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/init", `{"file_name":"takeout.zip","size":10}`)
	service := &handlerServiceMock{initFn: func(InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
		return InitTakeoutUploadResultDto{}, errors.New("disk exploded")
	}}

	NewHandler(service, nil).InitUploadHandler(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	if message := decodeErrorMessage(t, recorder); message != i18n.GetMessage("TAKEOUT_UPLOAD_INIT_ERROR") {
		t.Fatalf("unexpected message %q", message)
	}
}

func TestCompleteUploadHandlerReturns409ForAlreadyCompletedUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = buildJSONRequest(http.MethodPost, "/takeout/upload/complete", `{"upload_id":"u1"}`)
	service := &handlerServiceMock{completeFn: func(CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
		return TakeoutImportResultDto{}, ErrUploadAlreadyCompleted
	}}

	NewHandler(service, nil).CompleteUploadHandler(ctx)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
	message := decodeErrorMessage(t, recorder)
	if message != i18n.GetMessage("TAKEOUT_UPLOAD_ALREADY_COMPLETED") || message == "" {
		t.Fatalf("unexpected message %q", message)
	}
}
