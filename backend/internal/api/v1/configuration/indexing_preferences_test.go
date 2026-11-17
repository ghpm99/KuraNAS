package configuration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
)

func newServiceWithStoredDocument(t *testing.T, initialDocument string) *Service {
	t.Helper()
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() { config.AppConfig.Lang = originalLang })

	storedDocument := initialDocument
	return newConfigurationServiceForTest(t, &serviceRepoMock{
		getSettingsDocumentFn: func(string) (string, error) {
			if storedDocument == "" {
				return "", sql.ErrNoRows
			}
			return storedDocument, nil
		},
		upsertSettingsFn: func(_ *sql.Tx, _ string, payload string) error {
			storedDocument = payload
			return nil
		},
	})
}

func validUpdateRequest(indexing IndexingSettingsRequest) UpdateSettingsRequest {
	return UpdateSettingsRequest{
		Indexing:   indexing,
		Players:    PlayerSettingsRequest{ImageSlideshowSeconds: 4},
		Appearance: AppearanceSettingsRequest{AccentColor: "violet", ThemeMode: "dark"},
		Language:   LanguageSettingsRequest{Current: "en-US"},
	}
}

func intPtr(value int) *int { return &value }

func TestIndexingSettingsDefaultForDocumentSavedBeforeNewKeys(t *testing.T) {
	legacyDocument := `{"indexing":{"scan_on_startup":true,"extract_metadata":true,"generate_previews":false}}`
	service := newServiceWithStoredDocument(t, legacyDocument)

	settings, err := service.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if !settings.Indexing.IndexDocumentText || settings.Indexing.ScanIntervalHours != 24 {
		t.Fatalf("expected defaults true/24, got %+v", settings.Indexing)
	}

	preferences := service.IndexingPreferences()
	if !preferences.IndexDocumentText || preferences.ScanInterval != 24*time.Hour {
		t.Fatalf("expected reader defaults true/24h, got %+v", preferences)
	}
	if preferences.GeneratePreviews {
		t.Fatalf("expected saved generate_previews=false to be kept")
	}
}

func TestUpdateSettingsPersistsNewIndexingKeys(t *testing.T) {
	service := newServiceWithStoredDocument(t, "")
	disabled := false

	updated, err := service.UpdateSettings(validUpdateRequest(IndexingSettingsRequest{
		ExtractMetadata:   true,
		GeneratePreviews:  true,
		IndexDocumentText: &disabled,
		ScanIntervalHours: intPtr(12),
	}))
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}
	if updated.Indexing.IndexDocumentText || updated.Indexing.ScanIntervalHours != 12 {
		t.Fatalf("expected false/12 in response, got %+v", updated.Indexing)
	}

	reloaded, err := service.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if reloaded.Indexing.IndexDocumentText || reloaded.Indexing.ScanIntervalHours != 12 {
		t.Fatalf("expected false/12 persisted, got %+v", reloaded.Indexing)
	}
}

func TestUpdateSettingsWithoutNewIndexingKeysKeepsSavedValues(t *testing.T) {
	service := newServiceWithStoredDocument(t, "")
	disabled := false
	if _, err := service.UpdateSettings(validUpdateRequest(IndexingSettingsRequest{
		IndexDocumentText: &disabled,
		ScanIntervalHours: intPtr(12),
	})); err != nil {
		t.Fatalf("first UpdateSettings returned error: %v", err)
	}

	updated, err := service.UpdateSettings(validUpdateRequest(IndexingSettingsRequest{ExtractMetadata: true}))
	if err != nil {
		t.Fatalf("second UpdateSettings returned error: %v", err)
	}
	if updated.Indexing.IndexDocumentText || updated.Indexing.ScanIntervalHours != 12 {
		t.Fatalf("expected saved false/12 kept, got %+v", updated.Indexing)
	}
}

func TestUpdateSettingsRejectsScanIntervalOutOfRange(t *testing.T) {
	for _, hours := range []int{0, 169, -1} {
		service := newServiceWithStoredDocument(t, "")

		_, err := service.UpdateSettings(validUpdateRequest(IndexingSettingsRequest{ScanIntervalHours: intPtr(hours)}))
		if !errors.Is(err, ErrScanIntervalOutOfRange) {
			t.Fatalf("expected ErrScanIntervalOutOfRange for %d, got %v", hours, err)
		}
	}
}

func TestUpdateSettingsHandlerRejectsScanIntervalOutOfRangeWithTranslatedMessage(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() { config.AppConfig.Lang = originalLang })
	config.AppConfig.Lang = "pt-BR"
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	service := newServiceWithStoredDocument(t, "")
	config.AppConfig.Lang = "pt-BR"
	handler := NewHandler(service, &loggerMock{})

	for _, hours := range []string{"0", "169"} {
		body := bytes.NewBufferString(`{"indexing":{"scan_interval_hours":` + hours + `},` +
			`"players":{"image_slideshow_seconds":4},"appearance":{"accent_color":"violet"},"language":{"current":"en-US"}}`)
		ctx, recorder := newTestContext(http.MethodPut, body)

		handler.UpdateSettingsHandler(ctx)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d: %s", hours, recorder.Code, recorder.Body.String())
		}
		var response map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("response is not JSON: %v", err)
		}
		expected := i18n.GetMessage("ERROR_SCAN_INTERVAL_OUT_OF_RANGE")
		if expected == "ERROR_SCAN_INTERVAL_OUT_OF_RANGE" || response["error"] != expected {
			t.Fatalf("expected translated message %q, got %q", expected, response["error"])
		}
	}
}

func TestUpdateSettingsHandlerDecodesNewIndexingKeys(t *testing.T) {
	var captured UpdateSettingsRequest
	handler := NewHandler(&serviceMock{
		updateSettingsFn: func(request UpdateSettingsRequest) (SettingsDto, error) {
			captured = request
			return SettingsDto{}, nil
		},
	}, &loggerMock{})

	body := bytes.NewBufferString(`{"indexing":{"index_document_text":false,"scan_interval_hours":48}}`)
	ctx, recorder := newTestContext(http.MethodPut, body)
	handler.UpdateSettingsHandler(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if captured.Indexing.IndexDocumentText == nil || *captured.Indexing.IndexDocumentText {
		t.Fatalf("index_document_text did not decode: %+v", captured.Indexing)
	}
	if captured.Indexing.ScanIntervalHours == nil || *captured.Indexing.ScanIntervalHours != 48 {
		t.Fatalf("scan_interval_hours did not decode: %+v", captured.Indexing)
	}

	absentBody := bytes.NewBufferString(`{"indexing":{"extract_metadata":true}}`)
	absentCtx, _ := newTestContext(http.MethodPut, absentBody)
	handler.UpdateSettingsHandler(absentCtx)
	if captured.Indexing.IndexDocumentText != nil || captured.Indexing.ScanIntervalHours != nil {
		t.Fatalf("absent keys must decode as nil: %+v", captured.Indexing)
	}
}

func TestIndexingPreferencesReflectUpdateImmediately(t *testing.T) {
	service := newServiceWithStoredDocument(t, "")
	if preferences := service.IndexingPreferences(); preferences != DefaultIndexingPreferences() {
		t.Fatalf("expected defaults without document, got %+v", preferences)
	}

	disabled := false
	if _, err := service.UpdateSettings(validUpdateRequest(IndexingSettingsRequest{
		ExtractMetadata:   true,
		GeneratePreviews:  false,
		IndexDocumentText: &disabled,
		ScanIntervalHours: intPtr(6),
	})); err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	expected := IndexingPreferences{ExtractMetadata: true, GeneratePreviews: false, IndexDocumentText: false, ScanInterval: 6 * time.Hour}
	if preferences := service.IndexingPreferences(); preferences != expected {
		t.Fatalf("expected %+v, got %+v", expected, preferences)
	}
}

func TestIndexingPreferencesFallBackToDefaultsOnReadError(t *testing.T) {
	service := newConfigurationServiceForTest(t, &serviceRepoMock{
		getSettingsDocumentFn: func(string) (string, error) { return "", sql.ErrConnDone },
	})
	if preferences := service.IndexingPreferences(); preferences != DefaultIndexingPreferences() {
		t.Fatalf("expected defaults on read error, got %+v", preferences)
	}

	corrupted := newConfigurationServiceForTest(t, &serviceRepoMock{
		getSettingsDocumentFn: func(string) (string, error) { return "{", nil },
	})
	if preferences := corrupted.IndexingPreferences(); preferences != DefaultIndexingPreferences() {
		t.Fatalf("expected defaults on corrupted document, got %+v", preferences)
	}
}
