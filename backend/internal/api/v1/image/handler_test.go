package image

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

var errTest = errors.New("test error")

type fakeImageService struct {
	count      int
	countErr   error
	jobID      int
	enqueueErr error
}

func (f *fakeImageService) GetImages(page, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{}, nil
}

func (f *fakeImageService) GetPendingAIClassificationCount() (int, error) {
	return f.count, f.countErr
}

func (f *fakeImageService) EnqueueClassificationBackfill() (int, error) {
	return f.jobID, f.enqueueErr
}

type imageLoggerMock struct{ logger.LoggerServiceInterface }

func (m *imageLoggerMock) CreateLog(l logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *imageLoggerMock) CompleteWithSuccessLog(l logger.LoggerModel) error          { return nil }
func (m *imageLoggerMock) CompleteWithErrorLog(l logger.LoggerModel, err error) error { return nil }

func newTestHandler(svc ServiceInterface) *Handler {
	gin.SetMode(gin.TestMode)
	return NewHandler(svc, &imageLoggerMock{})
}

func newTestContext(method string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", nil)
	return c, w
}

func TestGetPendingAIClassificationCountHandler_Success(t *testing.T) {
	h := newTestHandler(&fakeImageService{count: 9})
	c, w := newTestContext(http.MethodGet)

	h.GetPendingAIClassificationCountHandler(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["pending_count"] != 9 {
		t.Fatalf("expected pending_count 9, got %d", body["pending_count"])
	}
}

func TestGetPendingAIClassificationCountHandler_Error(t *testing.T) {
	h := newTestHandler(&fakeImageService{countErr: errTest})
	c, w := newTestContext(http.MethodGet)

	h.GetPendingAIClassificationCountHandler(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEnqueueClassificationBackfillHandler_Success(t *testing.T) {
	h := newTestHandler(&fakeImageService{jobID: 77})
	c, w := newTestContext(http.MethodPost)

	h.EnqueueClassificationBackfillHandler(c)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if int(body["job_id"].(float64)) != 77 {
		t.Fatalf("expected job_id 77, got %v", body["job_id"])
	}
}

func TestEnqueueClassificationBackfillHandler_Unavailable(t *testing.T) {
	h := newTestHandler(&fakeImageService{enqueueErr: ErrBackfillUnavailable})
	c, w := newTestContext(http.MethodPost)

	h.EnqueueClassificationBackfillHandler(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestEnqueueClassificationBackfillHandler_Error(t *testing.T) {
	h := newTestHandler(&fakeImageService{enqueueErr: errTest})
	c, w := newTestContext(http.MethodPost)

	h.EnqueueClassificationBackfillHandler(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
