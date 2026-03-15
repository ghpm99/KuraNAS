package configuration

import (
	"bytes"
	"errors"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type loggerMock struct {
	createCalled        bool
	completeCalled      bool
	completeErrorCalled bool
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
	m.completeErrorCalled = true
	return nil
}

type serviceMock struct {
	getSettingsFn          func() (SettingsDto, error)
	updateSettingsFn       func(request UpdateSettingsRequest) (SettingsDto, error)
	getTranslationFilePath func() (string, error)
}

func (m *serviceMock) GetSettings() (SettingsDto, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn()
	}
	return SettingsDto{}, nil
}

func (m *serviceMock) UpdateSettings(request UpdateSettingsRequest) (SettingsDto, error) {
	if m.updateSettingsFn != nil {
		return m.updateSettingsFn(request)
	}
	return SettingsDto{}, nil
}

func (m *serviceMock) GetTranslationFilePath() (string, error) {
	if m.getTranslationFilePath != nil {
		return m.getTranslationFilePath()
	}
	return i18n.GetPathFileTranslate(), nil
}

func (m *serviceMock) ApplyRuntimeSettings() error {
	return nil
}

func newTestContext(method string, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	var requestBody *bytes.Buffer
	if body != nil {
		requestBody = body
	} else {
		requestBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, "/", requestBody)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestNewHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{}, l)
	if h == nil {
		t.Fatalf("expected handler instance")
	}
}

func TestGetAboutHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{
		getSettingsFn: func() (SettingsDto, error) {
			return SettingsDto{
				Language: LanguageSettingsDto{Current: "pt-BR"},
			}, nil
		},
	}, l)
	ctx, rec := newTestContext(http.MethodGet, nil)

	config.AppConfig.EntryPoint = "/data"
	config.AppConfig.Lang = "en-US"
	config.AppConfig.EnableWorkers = true
	config.AppConfig.StartupTime = time.Now()

	h.GetAboutHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"lang":"pt-BR"`) {
		t.Fatalf("expected response to contain service language, got %s", rec.Body.String())
	}
	if !l.createCalled || !l.completeCalled {
		t.Fatalf("expected logger create and complete calls")
	}
}

func TestGetTranslationJson(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{
		getTranslationFilePath: func() (string, error) {
			return i18n.GetPathFileTranslateByLang("en-US"), nil
		},
	}, l)
	ctx, rec := newTestContext(http.MethodGet, nil)

	h.GetTranslationJson(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected translation response status 200, got %d", rec.Code)
	}
	if !l.createCalled || !l.completeCalled {
		t.Fatalf("expected logger calls in translation endpoint")
	}
}

func TestGetSettingsHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{
		getSettingsFn: func() (SettingsDto, error) {
			return SettingsDto{
				Language: LanguageSettingsDto{Current: "en-US", Available: []string{"en-US", "pt-BR"}},
			}, nil
		},
	}, l)
	ctx, rec := newTestContext(http.MethodGet, nil)

	h.GetSettingsHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"current":"en-US"`) {
		t.Fatalf("expected settings payload, got %s", rec.Body.String())
	}
}

func TestGetSettingsHandlerServiceError(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{
		getSettingsFn: func() (SettingsDto, error) {
			return SettingsDto{}, errors.New("load failed")
		},
	}, l)
	ctx, rec := newTestContext(http.MethodGet, nil)

	h.GetSettingsHandler(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !l.completeErrorCalled {
		t.Fatalf("expected error log completion")
	}
}

func TestUpdateSettingsHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{
		updateSettingsFn: func(request UpdateSettingsRequest) (SettingsDto, error) {
			if request.Language.Current != "en-US" {
				t.Fatalf("expected request language to be passed through")
			}
			return SettingsDto{
				Language: LanguageSettingsDto{Current: request.Language.Current},
			}, nil
		},
	}, l)

	body := bytes.NewBufferString(`{
		"library": {"watched_paths": ["/data"], "remember_last_location": true, "prioritize_favorites": true},
		"indexing": {"scan_on_startup": true, "extract_metadata": true, "generate_previews": true},
		"players": {"remember_music_queue": true, "remember_video_progress": true, "autoplay_next_video": true, "image_slideshow_seconds": 4},
		"appearance": {"accent_color": "violet", "reduce_motion": false},
		"language": {"current": "en-US"}
	}`)
	ctx, rec := newTestContext(http.MethodPut, body)

	h.UpdateSettingsHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"current":"en-US"`) {
		t.Fatalf("expected updated settings response, got %s", rec.Body.String())
	}
}

func TestUpdateSettingsHandlerInvalidJSON(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{}, l)
	ctx, rec := newTestContext(http.MethodPut, bytes.NewBufferString(`{`))

	h.UpdateSettingsHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
