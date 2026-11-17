package autoshutdown

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	getFn       func() (SettingsDto, error)
	updateFn    func(dto SettingsDto) (SettingsDto, error)
	suggestedFn func() (SuggestedTimeDto, error)
}

func (m *mockService) GetSettings() (SettingsDto, error)                   { return m.getFn() }
func (m *mockService) UpdateSettings(dto SettingsDto) (SettingsDto, error) { return m.updateFn(dto) }
func (m *mockService) SuggestedTime() (SuggestedTimeDto, error)            { return m.suggestedFn() }
func (m *mockService) DueNow(t time.Time) (bool, int, error)               { return false, 0, nil }

func newTestRouter(service ServiceInterface) *gin.Engine {
	return newTestRouterWithCountdown(service, NewCountdown())
}

func newTestRouterWithCountdown(service ServiceInterface, countdown CountdownInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service, countdown)
	router.GET("/auto-shutdown/status", handler.GetStatusHandler)
	router.POST("/auto-shutdown/cancel", handler.CancelHandler)
	router.GET("/auto-shutdown/settings", handler.GetSettingsHandler)
	router.PUT("/auto-shutdown/settings", handler.UpdateSettingsHandler)
	router.GET("/auto-shutdown/suggested-time", handler.GetSuggestedTimeHandler)
	return router
}

func performRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestGetSettingsHandlerOK(t *testing.T) {
	router := newTestRouter(&mockService{
		getFn: func() (SettingsDto, error) {
			return SettingsDto{Enabled: true, Time: "03:00", GracePeriodSeconds: 60}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/auto-shutdown/settings", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "03:00") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

// TestUpdateSettingsHandlerDecodesPayload pins the request seam: it proves the
// handler decodes the exact JSON the frontend service sends (service/autoShutdown.ts
// → PUT /auto-shutdown/settings) into the right SettingsDto fields, and echoes
// the saved settings back. A json tag drift fails here instead of breaking the
// frontend integration silently.
func TestUpdateSettingsHandlerDecodesPayload(t *testing.T) {
	var captured SettingsDto
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) {
			captured = dto
			return dto, nil
		},
	})

	body := `{"enabled":true,"time":"03:30","grace_period_seconds":120}`
	response := performRequest(router, http.MethodPut, "/auto-shutdown/settings", body)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	want := SettingsDto{Enabled: true, Time: "03:30", GracePeriodSeconds: 120}
	if captured != want {
		t.Fatalf("handler decoded payload into %#v, want %#v", captured, want)
	}

	if !strings.Contains(response.Body.String(), `"time":"03:30"`) ||
		!strings.Contains(response.Body.String(), `"grace_period_seconds":120`) {
		t.Fatalf("response did not echo the saved settings: %s", response.Body.String())
	}
}

func TestUpdateSettingsHandlerInvalid(t *testing.T) {
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) {
			return SettingsDto{}, ErrInvalidSettingsRequest
		},
	})

	response := performRequest(router, http.MethodPut, "/auto-shutdown/settings", `{"enabled":true,"time":"99:99"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestUpdateSettingsHandlerMalformedBody(t *testing.T) {
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) { t.Fatal("must not be called"); return dto, nil },
	})

	response := performRequest(router, http.MethodPut, "/auto-shutdown/settings", `{nope`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestGetSuggestedTimeHandlerOK(t *testing.T) {
	router := newTestRouter(&mockService{
		suggestedFn: func() (SuggestedTimeDto, error) {
			return SuggestedTimeDto{Available: true, Time: "02:30", SampleSize: 7}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/auto-shutdown/suggested-time", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "02:30") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func warningCountdown() *Countdown {
	countdown := NewCountdown()
	now := time.Now()
	countdown.nowFn = func() time.Time { return now }
	countdown.begin(now.Add(45 * time.Second))
	return countdown
}

func decodeStatus(t *testing.T, body string) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("status body is not JSON: %v (%s)", err, body)
	}
	return decoded
}

func TestGetStatusHandlerIdle(t *testing.T) {
	router := newTestRouter(&mockService{})

	response := performRequest(router, http.MethodGet, "/auto-shutdown/status", "")

	decoded := decodeStatus(t, response.Body.String())
	if response.Code != http.StatusOK || decoded["state"] != "idle" ||
		decoded["shutdown_at"] != nil || decoded["seconds_remaining"] != nil || decoded["postponed_reason"] != "" {
		t.Fatalf("unexpected idle status: %d %s", response.Code, response.Body.String())
	}
}

func TestGetStatusHandlerWarning(t *testing.T) {
	router := newTestRouterWithCountdown(&mockService{}, warningCountdown())

	response := performRequest(router, http.MethodGet, "/auto-shutdown/status", "")

	decoded := decodeStatus(t, response.Body.String())
	if decoded["state"] != "warning" || decoded["seconds_remaining"] != float64(45) {
		t.Fatalf("unexpected warning status: %s", response.Body.String())
	}
	if shutdownAt, ok := decoded["shutdown_at"].(string); !ok || shutdownAt == "" {
		t.Fatalf("expected shutdown_at timestamp, got %v", decoded["shutdown_at"])
	}
}

func TestCancelHandlerDuringWarningReturnsIdleStatus(t *testing.T) {
	countdown := warningCountdown()
	router := newTestRouterWithCountdown(&mockService{}, countdown)

	response := performRequest(router, http.MethodPost, "/auto-shutdown/cancel", "")

	decoded := decodeStatus(t, response.Body.String())
	if response.Code != http.StatusOK || decoded["state"] != "idle" {
		t.Fatalf("unexpected cancel response: %d %s", response.Code, response.Body.String())
	}
	if countdown.Status().State != "idle" {
		t.Fatal("expected the countdown to be idle after cancel")
	}
}

func TestCancelHandlerOutsideWarningReturnsTranslatedConflict(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() { config.AppConfig.Lang = originalLang })
	config.AppConfig.Lang = "pt-BR"
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}
	router := newTestRouter(&mockService{})

	response := performRequest(router, http.MethodPost, "/auto-shutdown/cancel", "")

	decoded := decodeStatus(t, response.Body.String())
	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
	if decoded["error"] != "Não há desligamento em contagem para cancelar." {
		t.Fatalf("expected translated message, got %v", decoded["error"])
	}
}

func TestUpdateSettingsHandlerGracePeriodOutOfRangeReturnsTranslatedMessage(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() { config.AppConfig.Lang = originalLang })
	config.AppConfig.Lang = "pt-BR"
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) { return SettingsDto{}, ErrGracePeriodOutOfRange },
	})

	response := performRequest(router, http.MethodPut, "/auto-shutdown/settings", `{"enabled":true,"time":"03:00","grace_period_seconds":29}`)

	decoded := decodeStatus(t, response.Body.String())
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if decoded["error"] != "O período de aviso deve ficar entre 30 e 600 segundos." {
		t.Fatalf("expected translated message, got %v", decoded["error"])
	}
}
