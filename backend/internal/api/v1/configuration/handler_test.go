package configuration

import (
	"errors"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type loggerMock struct {
	createCalled   bool
	completeCalled bool
}

func (m *loggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	m.createCalled = true
	return log, nil
}
func (m *loggerMock) GetLogByID(id int) (logger.LoggerModel, error) { return logger.LoggerModel{}, nil }
func (m *loggerMock) GetLogs(page, pageSize int) ([]logger.LoggerModel, error) {
	return nil, nil
}
func (m *loggerMock) UpdateLog(log logger.LoggerModel) error { return nil }
func (m *loggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error {
	m.completeCalled = true
	return nil
}
func (m *loggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return errors.New("not used")
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest("GET", "/", nil)
	ctx.Request = req
	return ctx, rec
}

func TestNewHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(l)
	if h == nil {
		t.Fatalf("expected handler instance")
	}
}

func TestGetAboutHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(l)
	ctx, rec := newTestContext()

	config.AppConfig.EntryPoint = "/data"
	config.AppConfig.Lang = "pt-BR"
	config.AppConfig.EnableWorkers = true
	config.AppConfig.StartupTime = time.Now()

	h.GetAboutHandler(ctx)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !l.createCalled || !l.completeCalled {
		t.Fatalf("expected logger create and complete calls")
	}
}

func TestGetTranslationJson(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(l)
	ctx, _ := newTestContext()

	h.GetTranslationJson(ctx)

	if !l.createCalled || !l.completeCalled {
		t.Fatalf("expected logger calls in translation endpoint")
	}
}
