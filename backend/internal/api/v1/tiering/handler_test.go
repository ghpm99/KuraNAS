package tiering

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tieringengine "nas-go/api/internal/worker/tiering"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	getFn     func() (SettingsDto, error)
	updateFn  func(dto SettingsDto) (SettingsDto, error)
	statusFn  func() (StatusDto, error)
	usageFn   func() (TierUsageDto, error)
	promoteFn func(fileID int) (FileLocationDto, error)
}

func (m *mockService) GetSettings() (SettingsDto, error)                   { return m.getFn() }
func (m *mockService) UpdateSettings(dto SettingsDto) (SettingsDto, error) { return m.updateFn(dto) }
func (m *mockService) Status() (StatusDto, error)                          { return m.statusFn() }
func (m *mockService) Usage() (TierUsageDto, error)                        { return m.usageFn() }
func (m *mockService) PromoteFile(fileID int) (FileLocationDto, error) {
	return m.promoteFn(fileID)
}
func (m *mockService) MigrationPlan(now time.Time) (bool, string, []tieringengine.Promotion, []tieringengine.Demotion, error) {
	return false, "", nil, nil, nil
}
func (m *mockService) SetPhysicalPath(fileID int, physicalPath string) error { return nil }
func (m *mockService) NextRunDue(now time.Time) (bool, error)                { return false, nil }

func newTestRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service)
	router.GET("/tiering/settings", handler.GetSettingsHandler)
	router.PUT("/tiering/settings", handler.UpdateSettingsHandler)
	router.GET("/tiering/status", handler.GetStatusHandler)
	router.GET("/tiering/usage", handler.GetUsageHandler)
	router.POST("/tiering/promote/:file_id", handler.PromoteFileHandler)
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
			return SettingsDto{Enabled: true, ColdDirPath: "/mnt/cold", MinAgeDays: 90, IntervalHours: 24}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/tiering/settings", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "/mnt/cold") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

// TestUpdateSettingsHandlerDecodesPayload pins the request seam: it proves the
// handler decodes the exact JSON the frontend service sends (service/tiering.ts
// → PUT /tiering/settings) into the right SettingsDto fields, and echoes the
// saved settings back. A json tag drift fails here instead of breaking the
// frontend integration silently.
func TestUpdateSettingsHandlerDecodesPayload(t *testing.T) {
	var captured SettingsDto
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) {
			captured = dto
			return dto, nil
		},
	})

	body := `{"enabled":true,"cold_dir_path":"/mnt/cold","min_age_days":90,"min_size_bytes":1048576,"interval_hours":12}`
	response := performRequest(router, http.MethodPut, "/tiering/settings", body)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	want := SettingsDto{Enabled: true, ColdDirPath: "/mnt/cold", MinAgeDays: 90, MinSizeBytes: 1048576, IntervalHours: 12}
	if captured != want {
		t.Fatalf("handler decoded payload into %#v, want %#v", captured, want)
	}

	if !strings.Contains(response.Body.String(), `"cold_dir_path":"/mnt/cold"`) ||
		!strings.Contains(response.Body.String(), `"min_size_bytes":1048576`) {
		t.Fatalf("response did not echo the saved settings: %s", response.Body.String())
	}
}

func TestUpdateSettingsHandlerInvalidColdDir(t *testing.T) {
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) {
			return SettingsDto{}, ErrInvalidColdDir
		},
	})

	response := performRequest(router, http.MethodPut, "/tiering/settings", `{"enabled":true,"cold_dir_path":"/mnt/dados/cold"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestUpdateSettingsHandlerMalformedBody(t *testing.T) {
	router := newTestRouter(&mockService{
		updateFn: func(dto SettingsDto) (SettingsDto, error) { t.Fatal("must not be called"); return dto, nil },
	})

	response := performRequest(router, http.MethodPut, "/tiering/settings", `{nope`)
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

	response := performRequest(router, http.MethodGet, "/tiering/status", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "completed") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestGetUsageHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		usageFn: func() (TierUsageDto, error) {
			return TierUsageDto{HotFiles: 10, ColdFiles: 3, ColdBytes: 4096}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/tiering/usage", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "4096") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestTieringHandlersServerErrors(t *testing.T) {
	boom := errors.New("boom")
	router := newTestRouter(&mockService{
		getFn:    func() (SettingsDto, error) { return SettingsDto{}, boom },
		updateFn: func(dto SettingsDto) (SettingsDto, error) { return SettingsDto{}, boom },
		statusFn: func() (StatusDto, error) { return StatusDto{}, boom },
		usageFn:  func() (TierUsageDto, error) { return TierUsageDto{}, boom },
	})

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/tiering/settings", ""},
		{http.MethodPut, "/tiering/settings", `{"enabled":true,"cold_dir_path":"/mnt/cold"}`},
		{http.MethodGet, "/tiering/status", ""},
		{http.MethodGet, "/tiering/usage", ""},
	}

	for _, tc := range cases {
		response := performRequest(router, tc.method, tc.path, tc.body)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("%s %s: expected 500, got %d", tc.method, tc.path, response.Code)
		}
	}
}

func TestPromoteFileHandlerDecodesFileIdAndReturnsLocation(t *testing.T) {
	var receivedFileID int
	router := newTestRouter(&mockService{
		promoteFn: func(fileID int) (FileLocationDto, error) {
			receivedFileID = fileID
			return FileLocationDto{FileID: fileID, Tier: "hot", DiskPath: "/mnt/dados/a.pdf"}, nil
		},
	})

	response := performRequest(router, http.MethodPost, "/tiering/promote/42", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", response.Code, response.Body.String())
	}
	if receivedFileID != 42 {
		t.Fatalf("expected service to receive file id 42, got %d", receivedFileID)
	}
	var location FileLocationDto
	if err := json.Unmarshal(response.Body.Bytes(), &location); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if location.FileID != 42 || location.Tier != "hot" || location.DiskPath != "/mnt/dados/a.pdf" {
		t.Fatalf("unexpected location %+v", location)
	}
}

func TestPromoteFileHandlerRejectsInvalidFileId(t *testing.T) {
	router := newTestRouter(&mockService{})

	for _, invalidID := range []string{"abc", "0", "-3"} {
		response := performRequest(router, http.MethodPost, "/tiering/promote/"+invalidID, "")
		if response.Code != http.StatusBadRequest {
			t.Fatalf("id %q: expected 400, got %d", invalidID, response.Code)
		}
	}
}

func TestPromoteFileHandlerMapsServiceErrorsToTranslatedStatuses(t *testing.T) {
	cases := []struct {
		name            string
		serviceError    error
		expectedStatus  int
		expectedMessage string
	}{
		{"not found", ErrFileNotFound, http.StatusNotFound, i18n.GetMessage("ERROR_FILE_NOT_FOUND")},
		{"already hot", ErrFileAlreadyHot, http.StatusConflict, i18n.GetMessage("ERROR_TIERING_FILE_ALREADY_HOT")},
		{"no hot space", fmt.Errorf("wrapped: %w", ErrInsufficientHotSpace), http.StatusInsufficientStorage, i18n.GetMessage("ERROR_TIERING_NO_HOT_SPACE")},
		{"cold unavailable", ErrColdCopyUnavailable, http.StatusServiceUnavailable, i18n.GetMessage("ERROR_TIERING_COLD_UNAVAILABLE")},
		{"unexpected", errors.New("boom"), http.StatusInternalServerError, i18n.GetMessage("ERROR_TIERING_PROMOTE")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(&mockService{
				promoteFn: func(fileID int) (FileLocationDto, error) { return FileLocationDto{}, tc.serviceError },
			})

			response := performRequest(router, http.MethodPost, "/tiering/promote/7", "")

			if response.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, response.Code)
			}
			var body map[string]string
			_ = json.Unmarshal(response.Body.Bytes(), &body)
			if body["error"] != tc.expectedMessage {
				t.Fatalf("expected message %q, got %q", tc.expectedMessage, body["error"])
			}
		})
	}
}
