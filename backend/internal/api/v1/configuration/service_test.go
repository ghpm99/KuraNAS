package configuration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type serviceRepoMock struct {
	db                    *database.DbContext
	getSettingsDocumentFn func(settingKey string) (string, error)
	upsertSettingsFn      func(tx *sql.Tx, settingKey string, payload string) error
}

func (m *serviceRepoMock) GetDbContext() *database.DbContext { return m.db }

func (m *serviceRepoMock) GetSettingsDocument(settingKey string) (string, error) {
	if m.getSettingsDocumentFn != nil {
		return m.getSettingsDocumentFn(settingKey)
	}
	return "", sql.ErrNoRows
}

func (m *serviceRepoMock) UpsertSettingsDocument(tx *sql.Tx, settingKey string, payload string) error {
	if m.upsertSettingsFn != nil {
		return m.upsertSettingsFn(tx, settingKey, payload)
	}
	return nil
}

func newConfigurationServiceForTest(t *testing.T, repo *serviceRepoMock) *Service {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo.db = database.NewDbContext(db)
	return &Service{Repository: repo}
}

func TestNewServiceReturnsConfiguredImplementation(t *testing.T) {
	repo := &serviceRepoMock{}

	service := NewService(repo)
	typedService, ok := service.(*Service)
	if !ok {
		t.Fatalf("expected *Service implementation, got %T", service)
	}
	if typedService.Repository != repo {
		t.Fatalf("expected repository to be preserved")
	}
}

func TestConfigurationServiceGetSettingsUsesDefaultsWhenMissing(t *testing.T) {
	originalEntryPoint := config.AppConfig.EntryPoint
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
		config.AppConfig.Lang = originalLang
	})

	config.AppConfig.EntryPoint = "/runtime"
	config.AppConfig.Lang = "pt-BR"

	service := newConfigurationServiceForTest(t, &serviceRepoMock{
		getSettingsDocumentFn: func(settingKey string) (string, error) {
			if settingKey != settingsStorageKey {
				t.Fatalf("unexpected setting key %s", settingKey)
			}
			return "", sql.ErrNoRows
		},
	})

	settings, err := service.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if settings.Library.RuntimeRootPath != "/runtime" {
		t.Fatalf("expected runtime root path to match config")
	}
	if settings.Language.Current != "pt-BR" {
		t.Fatalf("expected current locale from config, got %s", settings.Language.Current)
	}
	if settings.Players.ImageSlideshowSeconds != 4 {
		t.Fatalf("expected default slideshow seconds, got %d", settings.Players.ImageSlideshowSeconds)
	}
}

func TestConfigurationServiceUpdateSettingsPersistsNormalizedState(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() {
		config.AppConfig.Lang = originalLang
	})

	var storedPayload string
	service := newConfigurationServiceForTest(t, &serviceRepoMock{
		upsertSettingsFn: func(tx *sql.Tx, settingKey string, payload string) error {
			if settingKey != settingsStorageKey {
				t.Fatalf("unexpected setting key %s", settingKey)
			}
			storedPayload = payload
			return nil
		},
	})

	settings, err := service.UpdateSettings(UpdateSettingsRequest{
		Library: LibrarySettingsRequest{
			WatchedPaths:         []string{"/data", "/data"},
			RememberLastLocation: true,
			PrioritizeFavorites:  false,
		},
		Indexing: IndexingSettingsRequest{
			ScanOnStartup:    false,
			ExtractMetadata:  true,
			GeneratePreviews: false,
		},
		Players: PlayerSettingsRequest{
			RememberMusicQueue:    true,
			RememberVideoProgress: true,
			AutoplayNextVideo:     false,
			ImageSlideshowSeconds: 8,
		},
		Appearance: AppearanceSettingsRequest{
			AccentColor:  "cyan",
			ReduceMotion: true,
		},
		Language: LanguageSettingsRequest{
			Current: "en-US",
		},
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	if settings.Appearance.AccentColor != "cyan" {
		t.Fatalf("expected accent color to be persisted")
	}
	if settings.Players.ImageSlideshowSeconds != 8 {
		t.Fatalf("expected slideshow seconds to be persisted")
	}
	if config.AppConfig.Lang != "en-US" {
		t.Fatalf("expected runtime language to be updated")
	}
	if storedPayload == "" {
		t.Fatalf("expected serialized payload to be stored")
	}
}

func TestConfigurationServiceUpdateSettingsValidatesRequest(t *testing.T) {
	service := newConfigurationServiceForTest(t, &serviceRepoMock{})

	_, err := service.UpdateSettings(UpdateSettingsRequest{
		Players: PlayerSettingsRequest{
			ImageSlideshowSeconds: 1,
		},
		Appearance: AppearanceSettingsRequest{
			AccentColor: "unknown",
		},
		Language: LanguageSettingsRequest{
			Current: "de-DE",
		},
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !errors.Is(err, ErrInvalidSettingsRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}

func TestConfigurationServiceGetTranslationFilePathFallsBackOnError(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() {
		config.AppConfig.Lang = originalLang
	})
	config.AppConfig.Lang = "en-US"

	service := newConfigurationServiceForTest(t, &serviceRepoMock{
		getSettingsDocumentFn: func(settingKey string) (string, error) {
			return "", errors.New("read failed")
		},
	})

	path, err := service.GetTranslationFilePath()
	if err == nil {
		t.Fatalf("expected error to be returned")
	}
	if path == "" {
		t.Fatalf("expected fallback translation path")
	}
}

func TestConfigurationServiceApplyRuntimeSettingsUsesStoredLocale(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() {
		config.AppConfig.Lang = originalLang
	})
	config.AppConfig.Lang = "pt-BR"

	payload, err := json.Marshal(settingsState{
		Language: languageSettingsState{
			Current: "en-US",
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal stored settings: %v", err)
	}

	service := newConfigurationServiceForTest(t, &serviceRepoMock{
		getSettingsDocumentFn: func(settingKey string) (string, error) {
			if settingKey != settingsStorageKey {
				t.Fatalf("unexpected setting key %s", settingKey)
			}
			return string(payload), nil
		},
	})

	if err := service.ApplyRuntimeSettings(); err != nil {
		t.Fatalf("ApplyRuntimeSettings returned error: %v", err)
	}
	if config.AppConfig.Lang != "en-US" {
		t.Fatalf("expected runtime language to be updated, got %s", config.AppConfig.Lang)
	}
}
