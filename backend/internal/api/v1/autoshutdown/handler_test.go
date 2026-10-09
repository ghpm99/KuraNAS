package autoshutdown

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service)
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
