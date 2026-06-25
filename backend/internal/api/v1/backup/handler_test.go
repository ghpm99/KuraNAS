package backup

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	backupengine "nas-go/api/internal/worker/backup"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	getFn     func() (SettingsDto, error)
	updateFn  func(dto SettingsDto) (SettingsDto, error)
	statusFn  func() (StatusDto, error)
	pendingFn func() (PendingDto, error)
}

func (m *mockService) GetSettings() (SettingsDto, error)                   { return m.getFn() }
func (m *mockService) UpdateSettings(dto SettingsDto) (SettingsDto, error) { return m.updateFn(dto) }
func (m *mockService) Status() (StatusDto, error)                          { return m.statusFn() }
func (m *mockService) Pending() (PendingDto, error)                        { return m.pendingFn() }
func (m *mockService) RunOptions() (bool, backupengine.Options, error) {
	return false, backupengine.Options{}, nil
}
func (m *mockService) NextRunDue(now time.Time) (bool, error) { return false, nil }

func newTestRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service)
	router.GET("/backup/settings", handler.GetSettingsHandler)
	router.PUT("/backup/settings", handler.UpdateSettingsHandler)
	router.GET("/backup/status", handler.GetStatusHandler)
	router.GET("/backup/pending", handler.GetPendingHandler)
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

func TestGetSettingsHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		getFn: func() (SettingsDto, error) {
			return SettingsDto{Enabled: true, DestinationPath: "/mnt/backup", RetentionDays: 30, IntervalHours: 24}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/backup/settings", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "/mnt/backup") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

// TestUpdateSettingsHandlerDecodesPayload pins the request seam: it proves the
// handler decodes the exact JSON the frontend service sends (service/backup.ts
// → PUT /backup/settings) into the right SettingsDto fields, and echoes the
// saved settings back. If a json tag drifts on either side, `captured` loses a
// field and this fails — instead of the integration breaking silently in prod.
func TestUpdateSettingsHandlerDecodesPayload(t *testing.T) {
	var captured SettingsDto
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) {
			captured = dto
			return dto, nil
		},
	})

	body := `{"enabled":true,"destination_path":"/mnt/cold","retention_days":15,"interval_hours":12}`
	response := performRequest(router, http.MethodPut, "/backup/settings", body)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	want := SettingsDto{Enabled: true, DestinationPath: "/mnt/cold", RetentionDays: 15, IntervalHours: 12}
	if captured != want {
		t.Fatalf("handler decoded payload into %#v, want %#v", captured, want)
	}

	if !strings.Contains(response.Body.String(), `"destination_path":"/mnt/cold"`) ||
		!strings.Contains(response.Body.String(), `"interval_hours":12`) {
		t.Fatalf("response did not echo the saved settings: %s", response.Body.String())
	}
}

func TestUpdateSettingsHandlerInvalidDestination(t *testing.T) {
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) {
			return SettingsDto{}, ErrInvalidDestination
		},
	})

	response := performRequest(router, http.MethodPut, "/backup/settings", `{"enabled":true,"destination_path":"/mnt/dados/backup"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestUpdateSettingsHandlerMalformedBody(t *testing.T) {
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) { t.Fatal("must not be called"); return dto, nil },
	})

	response := performRequest(router, http.MethodPut, "/backup/settings", `{nope`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestGetStatusHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		statusFn: func() (StatusDto, error) {
			return StatusDto{Enabled: true, HasRun: true, Status: "completed"}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/backup/status", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "completed") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestGetPendingHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		pendingFn: func() (PendingDto, error) { return PendingDto{PendingFiles: 5}, nil },
	})

	response := performRequest(router, http.MethodGet, "/backup/pending", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "5") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestBackupHandlersServerErrors(t *testing.T) {
	boom := errors.New("boom")
	router := newTestRouter(&mockService{
		getFn:     func() (SettingsDto, error) { return SettingsDto{}, boom },
		updateFn:  func(dto SettingsDto) (SettingsDto, error) { return SettingsDto{}, boom },
		statusFn:  func() (StatusDto, error) { return StatusDto{}, boom },
		pendingFn: func() (PendingDto, error) { return PendingDto{}, boom },
	})

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/backup/settings", ""},
		{http.MethodPut, "/backup/settings", `{"enabled":true,"destination_path":"/mnt/backup"}`},
		{http.MethodGet, "/backup/status", ""},
		{http.MethodGet, "/backup/pending", ""},
	}

	for _, tc := range cases {
		response := performRequest(router, tc.method, tc.path, tc.body)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("%s %s: expected 500, got %d", tc.method, tc.path, response.Code)
		}
	}
}
