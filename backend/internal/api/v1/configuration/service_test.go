package configuration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"testing"
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
	repo.db = database.NewDbContext(nil)
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

func TestResolveLocaleAllBranches(t *testing.T) {
	originalLang := config.AppConfig.Lang
	t.Cleanup(func() {
		config.AppConfig.Lang = originalLang
	})

	t.Run("uses value when available", func(t *testing.T) {
		config.AppConfig.Lang = ""
		got := resolveLocale("pt-BR", []string{"en-US", "pt-BR"})
		if got != "pt-BR" {
			t.Fatalf("expected pt-BR, got %s", got)
		}
	})

	t.Run("falls back to config when value empty", func(t *testing.T) {
		config.AppConfig.Lang = "pt-BR"
		got := resolveLocale("", []string{"en-US", "pt-BR"})
		if got != "pt-BR" {
			t.Fatalf("expected pt-BR from config, got %s", got)
		}
	})

	t.Run("falls back to default when both empty", func(t *testing.T) {
		config.AppConfig.Lang = ""
		got := resolveLocale("", []string{"en-US", "fr-FR"})
		if got != "en-US" {
			t.Fatalf("expected en-US default, got %s", got)
		}
	})

	t.Run("returns value when no available locales", func(t *testing.T) {
		config.AppConfig.Lang = ""
		got := resolveLocale("ja-JP", nil)
		if got != "ja-JP" {
			t.Fatalf("expected ja-JP, got %s", got)
		}
	})

	t.Run("falls back to default locale when value not in list", func(t *testing.T) {
		config.AppConfig.Lang = ""
		got := resolveLocale("ja-JP", []string{"fr-FR", "en-US"})
		if got != "en-US" {
			t.Fatalf("expected en-US fallback, got %s", got)
		}
	})

	t.Run("returns first available when nothing matches", func(t *testing.T) {
		config.AppConfig.Lang = ""
		got := resolveLocale("ja-JP", []string{"fr-FR", "de-DE"})
		if got != "fr-FR" {
			t.Fatalf("expected first available fr-FR, got %s", got)
		}
	})
}

func TestSanitizePathsBranches(t *testing.T) {
	t.Run("deduplicates and trims", func(t *testing.T) {
		result := sanitizePaths([]string{" /a ", "/b", "/a", " "}, []string{"/default"})
		if len(result) != 2 || result[0] != "/a" || result[1] != "/b" {
			t.Fatalf("expected [/a /b], got %v", result)
		}
	})

	t.Run("returns fallback when all empty", func(t *testing.T) {
		result := sanitizePaths([]string{" ", ""}, []string{"/default"})
		if len(result) != 1 || result[0] != "/default" {
			t.Fatalf("expected [/default], got %v", result)
		}
	})

	t.Run("returns fallback when nil", func(t *testing.T) {
		result := sanitizePaths(nil, []string{"/d"})
		if len(result) != 1 || result[0] != "/d" {
			t.Fatalf("expected [/d], got %v", result)
		}
	})
}

func TestNormalizeSlideshowSeconds(t *testing.T) {
	if got := normalizeSlideshowSeconds(8, 4); got != 8 {
		t.Fatalf("expected 8, got %d", got)
	}
	if got := normalizeSlideshowSeconds(5, 4); got != 4 {
		t.Fatalf("expected fallback 4, got %d", got)
	}
}

func TestNormalizeAccentColor(t *testing.T) {
	if got := normalizeAccentColor("cyan", "violet"); got != "cyan" {
		t.Fatalf("expected cyan, got %s", got)
	}
	if got := normalizeAccentColor("blue", "violet"); got != "violet" {
		t.Fatalf("expected fallback violet, got %s", got)
	}
}

func TestValidateUpdateRequestBranches(t *testing.T) {
	validRequest := UpdateSettingsRequest{
		Appearance: AppearanceSettingsRequest{AccentColor: "violet"},
		Players:    PlayerSettingsRequest{ImageSlideshowSeconds: 4},
		Language:   LanguageSettingsRequest{Current: "en-US"},
	}

	t.Run("valid request", func(t *testing.T) {
		err := validateUpdateRequest(validRequest, []string{"en-US"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invalid accent color", func(t *testing.T) {
		req := validRequest
		req.Appearance.AccentColor = "red"
		err := validateUpdateRequest(req, []string{"en-US"})
		if !errors.Is(err, ErrInvalidSettingsRequest) {
			t.Fatalf("expected ErrInvalidSettingsRequest, got %v", err)
		}
	})

	t.Run("invalid slideshow seconds", func(t *testing.T) {
		req := validRequest
		req.Players.ImageSlideshowSeconds = 99
		err := validateUpdateRequest(req, []string{"en-US"})
		if !errors.Is(err, ErrInvalidSettingsRequest) {
			t.Fatalf("expected ErrInvalidSettingsRequest, got %v", err)
		}
	})

	t.Run("invalid locale", func(t *testing.T) {
		config.AppConfig.Lang = ""
		req := validRequest
		req.Language.Current = "xx-XX"
		err := validateUpdateRequest(req, []string{"en-US", "pt-BR"})
		if !errors.Is(err, ErrInvalidSettingsRequest) {
			t.Fatalf("expected ErrInvalidSettingsRequest, got %v", err)
		}
	})
}
