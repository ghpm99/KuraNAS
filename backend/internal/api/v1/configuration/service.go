package configuration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ErrInvalidSettingsRequest = errors.New("invalid settings request")

// ErrCapturesPathInsideRoot is returned when the captures save path would land
// inside a storage root (or a subfolder of one). Captures must live outside the
// indexed roots so they — and their in-progress upload staging — are never
// watched/indexed.
var ErrCapturesPathInsideRoot = errors.New("captures path must be outside every storage root")

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

func (s *Service) GetSettings() (SettingsDto, error) {
	state, availableLocales, err := s.loadState()
	if err != nil {
		return SettingsDto{}, err
	}

	return state.toDto(availableLocales), nil
}

func (s *Service) UpdateSettings(request UpdateSettingsRequest) (SettingsDto, error) {
	availableLocales, err := s.listAvailableLocales()
	if err != nil {
		return SettingsDto{}, err
	}

	if err := validateUpdateRequest(request, availableLocales); err != nil {
		return SettingsDto{}, err
	}

	defaults := buildDefaultSettings(availableLocales)
	normalized := normalizeState(request.toState(), defaults, availableLocales)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return SettingsDto{}, fmt.Errorf("falha ao serializar configuracoes: %w", err)
	}

	if err := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpsertSettingsDocument(tx, settingsStorageKey, string(payload))
	}); err != nil {
		return SettingsDto{}, fmt.Errorf("falha ao atualizar configuracoes: %w", err)
	}

	if err := applyRuntimeLanguage(normalized.Language.Current); err != nil {
		return SettingsDto{}, fmt.Errorf("falha ao aplicar idioma em runtime: %w", err)
	}
	applyRuntimeCapturesPath(normalized.Captures.SavePath)

	return normalized.toDto(availableLocales), nil
}

func (s *Service) GetTranslationFilePath() (string, error) {
	settings, err := s.GetSettings()
	if err != nil {
		return i18n.GetPathFileTranslateByLang(resolveLocale(config.AppConfig.Lang, nil)), err
	}

	return i18n.GetPathFileTranslateByLang(settings.Language.Current), nil
}

func (s *Service) ApplyRuntimeSettings() error {
	settings, err := s.GetSettings()
	if err != nil {
		return err
	}
	applyRuntimeCapturesPath(settings.Captures.SavePath)
	return applyRuntimeLanguage(settings.Language.Current)
}

// IsAIImageClassificationEnabled reports whether the worker may call AI to
// classify images. It reads only the stored document (a single indexed lookup,
// no translation listing) so it stays cheap enough to consult per file. A
// missing document or missing AI section resolves to the safe default (enabled).
func (s *Service) IsAIImageClassificationEnabled() (bool, error) {
	payload, err := s.Repository.GetSettingsDocument(settingsStorageKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultAIImageClassification, nil
		}
		return defaultAIImageClassification, fmt.Errorf("falha ao carregar configuracoes de IA: %w", err)
	}

	var stored settingsState
	if err := json.Unmarshal([]byte(payload), &stored); err != nil {
		return defaultAIImageClassification, fmt.Errorf("falha ao desserializar configuracoes de IA: %w", err)
	}

	return derefBool(stored.AI.ImageClassification, defaultAIImageClassification), nil
}

func (s *Service) loadState() (settingsState, []string, error) {
	availableLocales, err := s.listAvailableLocales()
	if err != nil {
		return settingsState{}, nil, err
	}

	defaults := buildDefaultSettings(availableLocales)
	payload, err := s.Repository.GetSettingsDocument(settingsStorageKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaults, availableLocales, nil
		}
		return settingsState{}, nil, fmt.Errorf("falha ao carregar configuracoes: %w", err)
	}

	var stored settingsState
	if err := json.Unmarshal([]byte(payload), &stored); err != nil {
		return settingsState{}, nil, fmt.Errorf("falha ao desserializar configuracoes: %w", err)
	}

	return normalizeState(stored, defaults, availableLocales), availableLocales, nil
}

func (s *Service) listAvailableLocales() ([]string, error) {
	translationsPath := i18n.ResolveTranslationsPath()
	entries, err := os.ReadDir(translationsPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar traducoes disponiveis: %w", err)
	}

	locales := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		locales = append(locales, strings.TrimSuffix(entry.Name(), ".json"))
	}

	if len(locales) == 0 {
		locales = append(locales, defaultLocale)
	}

	sort.Strings(locales)
	return locales, nil
}

func validateUpdateRequest(request UpdateSettingsRequest, availableLocales []string) error {
	if _, ok := allowedAccentColors[request.Appearance.AccentColor]; !ok {
		return fmt.Errorf("%w: accent_color", ErrInvalidSettingsRequest)
	}
	if _, ok := allowedSlideshowSeconds[request.Players.ImageSlideshowSeconds]; !ok {
		return fmt.Errorf("%w: image_slideshow_seconds", ErrInvalidSettingsRequest)
	}
	if resolveLocale(request.Language.Current, availableLocales) != strings.TrimSpace(request.Language.Current) {
		return fmt.Errorf("%w: language.current", ErrInvalidSettingsRequest)
	}
	if err := validateCapturesPath(request.Captures.SavePath); err != nil {
		return err
	}

	return nil
}

// validateCapturesPath enforces that a non-empty captures save path is absolute
// and lives OUTSIDE every storage root (an empty path is normalized to the
// default later). Rejecting in-root paths is the whole point of the setting: a
// captures folder inside a root would be watched/indexed, and the in-progress
// upload would be re-indexed on every chunk.
func validateCapturesPath(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	clean := filepath.Clean(trimmed)
	if !filepath.IsAbs(clean) {
		return fmt.Errorf("%w: captures.save_path must be absolute", ErrInvalidSettingsRequest)
	}
	if _, inside := roots.OwnerOf(clean); inside {
		return fmt.Errorf("%w (%s)", ErrCapturesPathInsideRoot, clean)
	}

	return nil
}

func applyRuntimeLanguage(locale string) error {
	config.AppConfig.Lang = locale
	return i18n.LoadTranslations()
}
