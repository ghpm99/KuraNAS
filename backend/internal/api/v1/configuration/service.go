package configuration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"os"
	"sort"
	"strings"
)

var ErrInvalidSettingsRequest = errors.New("invalid settings request")

type ServiceInterface interface {
	GetSettings() (SettingsDto, error)
	UpdateSettings(request UpdateSettingsRequest) (SettingsDto, error)
	GetTranslationFilePath() (string, error)
	ApplyRuntimeSettings() error
}

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
	return applyRuntimeLanguage(settings.Language.Current)
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

	return nil
}

func applyRuntimeLanguage(locale string) error {
	config.AppConfig.Lang = locale
	return i18n.LoadTranslations()
}
