package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/pkg/i18n"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type updaterServiceMock struct {
	checkFn    func() (UpdateStatusDto, error)
	downloadFn func() error
}

func (m *updaterServiceMock) CheckForUpdate() (UpdateStatusDto, error) { return m.checkFn() }
func (m *updaterServiceMock) DownloadAndApply() error                  { return m.downloadFn() }

type updaterLoggerMock struct{ logger.LoggerServiceInterface }

func (m *updaterLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *updaterLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (m *updaterLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func TestUpdaterHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	successService := &updaterServiceMock{
		checkFn: func() (UpdateStatusDto, error) {
			return UpdateStatusDto{UpdateAvailable: true, CurrentVersion: "1.0.0", LatestVersion: "1.1.0"}, nil
		},
		downloadFn: func() error { return nil },
	}
	errorService := &updaterServiceMock{
		checkFn:    func() (UpdateStatusDto, error) { return UpdateStatusDto{}, errors.New("check failed") },
		downloadFn: func() error { return errors.New("apply failed") },
	}

	r1 := gin.New()
	h1 := NewHandler(successService, &updaterLoggerMock{})
	r1.GET("/update/status", h1.GetUpdateStatusHandler)
	r1.POST("/update/apply", h1.ApplyUpdateHandler)

	w := httptest.NewRecorder()
	r1.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/update/status", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for status success, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r1.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/update/apply", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for apply success, got %d", w.Code)
	}

	r2 := gin.New()
	h2 := NewHandler(errorService, &updaterLoggerMock{})
	r2.GET("/update/status", h2.GetUpdateStatusHandler)
	r2.POST("/update/apply", h2.ApplyUpdateHandler)

	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/update/status", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 for status error, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/update/apply", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 for apply error, got %d", w.Code)
	}
}

func TestApplyUpdateHandlerRefusalsRespondConflictWithTranslatedMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("failed to load translations: %v", err)
	}
	cases := []struct {
		name       string
		refusal    error
		messageKey string
	}{
		{"missing checksum", errChecksumMissing, "ERROR_UPDATE_CHECKSUM_MISSING"},
		{"no entry", errChecksumNoEntry, "ERROR_UPDATE_CHECKSUM_NO_ENTRY"},
		{"mismatch", errChecksumMismatch, "ERROR_UPDATE_CHECKSUM_MISMATCH"},
		{"too large", errPackageTooLarge, "ERROR_UPDATE_PACKAGE_TOO_LARGE"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &updaterServiceMock{downloadFn: func() error { return fmt.Errorf("failed to verify: %w", tc.refusal) }}
			router := gin.New()
			router.POST("/update/apply", NewHandler(service, &updaterLoggerMock{}).ApplyUpdateHandler)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/update/apply", nil))

			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid body: %v", err)
			}
			if recorder.Code != http.StatusConflict || body.Error == tc.messageKey || body.Error != i18n.GetMessage(tc.messageKey) {
				t.Fatalf("expected 409 with %s, got %d %q", tc.messageKey, recorder.Code, body.Error)
			}
		})
	}
}
