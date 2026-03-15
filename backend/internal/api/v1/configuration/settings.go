package configuration

import (
	"nas-go/api/internal/config"
	"strings"
)

const (
	settingsStorageKey = "system_preferences"
	defaultLocale      = "en-US"
	defaultAccentColor = "violet"
)

var allowedSlideshowSeconds = map[int]struct{}{
	4:  {},
	8:  {},
	12: {},
	20: {},
}

var allowedAccentColors = map[string]struct{}{
	"violet": {},
	"cyan":   {},
	"rose":   {},
}

type SettingsDto struct {
	Library    LibrarySettingsDto    `json:"library"`
	Indexing   IndexingSettingsDto   `json:"indexing"`
	Players    PlayerSettingsDto     `json:"players"`
	Appearance AppearanceSettingsDto `json:"appearance"`
	Language   LanguageSettingsDto   `json:"language"`
}

type LibrarySettingsDto struct {
	RuntimeRootPath      string   `json:"runtime_root_path"`
	WatchedPaths         []string `json:"watched_paths"`
	RememberLastLocation bool     `json:"remember_last_location"`
	PrioritizeFavorites  bool     `json:"prioritize_favorites"`
}

type IndexingSettingsDto struct {
	WorkersEnabled   bool `json:"workers_enabled"`
	ScanOnStartup    bool `json:"scan_on_startup"`
	ExtractMetadata  bool `json:"extract_metadata"`
	GeneratePreviews bool `json:"generate_previews"`
}

type PlayerSettingsDto struct {
	RememberMusicQueue    bool `json:"remember_music_queue"`
	RememberVideoProgress bool `json:"remember_video_progress"`
	AutoplayNextVideo     bool `json:"autoplay_next_video"`
	ImageSlideshowSeconds int  `json:"image_slideshow_seconds"`
}

type AppearanceSettingsDto struct {
	AccentColor  string `json:"accent_color"`
	ReduceMotion bool   `json:"reduce_motion"`
}

type LanguageSettingsDto struct {
	Current   string   `json:"current"`
	Available []string `json:"available"`
}

type UpdateSettingsRequest struct {
	Library    LibrarySettingsRequest    `json:"library"`
	Indexing   IndexingSettingsRequest   `json:"indexing"`
	Players    PlayerSettingsRequest     `json:"players"`
	Appearance AppearanceSettingsRequest `json:"appearance"`
	Language   LanguageSettingsRequest   `json:"language"`
}

type LibrarySettingsRequest struct {
	WatchedPaths         []string `json:"watched_paths"`
	RememberLastLocation bool     `json:"remember_last_location"`
	PrioritizeFavorites  bool     `json:"prioritize_favorites"`
}

type IndexingSettingsRequest struct {
	ScanOnStartup    bool `json:"scan_on_startup"`
	ExtractMetadata  bool `json:"extract_metadata"`
	GeneratePreviews bool `json:"generate_previews"`
}

type PlayerSettingsRequest struct {
	RememberMusicQueue    bool `json:"remember_music_queue"`
	RememberVideoProgress bool `json:"remember_video_progress"`
	AutoplayNextVideo     bool `json:"autoplay_next_video"`
	ImageSlideshowSeconds int  `json:"image_slideshow_seconds"`
}

type AppearanceSettingsRequest struct {
	AccentColor  string `json:"accent_color"`
	ReduceMotion bool   `json:"reduce_motion"`
}

type LanguageSettingsRequest struct {
	Current string `json:"current"`
}

type settingsState struct {
	Library    librarySettingsState    `json:"library"`
	Indexing   indexingSettingsState   `json:"indexing"`
	Players    playerSettingsState     `json:"players"`
	Appearance appearanceSettingsState `json:"appearance"`
	Language   languageSettingsState   `json:"language"`
}

type librarySettingsState struct {
	WatchedPaths         []string `json:"watched_paths"`
	RememberLastLocation bool     `json:"remember_last_location"`
	PrioritizeFavorites  bool     `json:"prioritize_favorites"`
}

type indexingSettingsState struct {
	ScanOnStartup    bool `json:"scan_on_startup"`
	ExtractMetadata  bool `json:"extract_metadata"`
	GeneratePreviews bool `json:"generate_previews"`
}

type playerSettingsState struct {
	RememberMusicQueue    bool `json:"remember_music_queue"`
	RememberVideoProgress bool `json:"remember_video_progress"`
	AutoplayNextVideo     bool `json:"autoplay_next_video"`
	ImageSlideshowSeconds int  `json:"image_slideshow_seconds"`
}

type appearanceSettingsState struct {
	AccentColor  string `json:"accent_color"`
	ReduceMotion bool   `json:"reduce_motion"`
}

type languageSettingsState struct {
	Current string `json:"current"`
}

func buildDefaultSettings(availableLocales []string) settingsState {
	runtimeRootPath := strings.TrimSpace(config.AppConfig.EntryPoint)
	watchedPaths := []string{}
	if runtimeRootPath != "" {
		watchedPaths = []string{runtimeRootPath}
	}

	return settingsState{
		Library: librarySettingsState{
			WatchedPaths:         watchedPaths,
			RememberLastLocation: true,
			PrioritizeFavorites:  true,
		},
		Indexing: indexingSettingsState{
			ScanOnStartup:    true,
			ExtractMetadata:  true,
			GeneratePreviews: true,
		},
		Players: playerSettingsState{
			RememberMusicQueue:    true,
			RememberVideoProgress: true,
			AutoplayNextVideo:     true,
			ImageSlideshowSeconds: 4,
		},
		Appearance: appearanceSettingsState{
			AccentColor:  defaultAccentColor,
			ReduceMotion: false,
		},
		Language: languageSettingsState{
			Current: resolveLocale(config.AppConfig.Lang, availableLocales),
		},
	}
}

func normalizeState(candidate settingsState, defaults settingsState, availableLocales []string) settingsState {
	normalized := defaults
	normalized.Library.WatchedPaths = sanitizePaths(candidate.Library.WatchedPaths, defaults.Library.WatchedPaths)
	normalized.Library.RememberLastLocation = candidate.Library.RememberLastLocation
	normalized.Library.PrioritizeFavorites = candidate.Library.PrioritizeFavorites
	normalized.Indexing.ScanOnStartup = candidate.Indexing.ScanOnStartup
	normalized.Indexing.ExtractMetadata = candidate.Indexing.ExtractMetadata
	normalized.Indexing.GeneratePreviews = candidate.Indexing.GeneratePreviews
	normalized.Players.RememberMusicQueue = candidate.Players.RememberMusicQueue
	normalized.Players.RememberVideoProgress = candidate.Players.RememberVideoProgress
	normalized.Players.AutoplayNextVideo = candidate.Players.AutoplayNextVideo
	normalized.Players.ImageSlideshowSeconds = normalizeSlideshowSeconds(candidate.Players.ImageSlideshowSeconds, defaults.Players.ImageSlideshowSeconds)
	normalized.Appearance.AccentColor = normalizeAccentColor(candidate.Appearance.AccentColor, defaults.Appearance.AccentColor)
	normalized.Appearance.ReduceMotion = candidate.Appearance.ReduceMotion
	normalized.Language.Current = resolveLocale(candidate.Language.Current, availableLocales)
	return normalized
}

func (request UpdateSettingsRequest) toState() settingsState {
	return settingsState{
		Library: librarySettingsState{
			WatchedPaths:         request.Library.WatchedPaths,
			RememberLastLocation: request.Library.RememberLastLocation,
			PrioritizeFavorites:  request.Library.PrioritizeFavorites,
		},
		Indexing: indexingSettingsState{
			ScanOnStartup:    request.Indexing.ScanOnStartup,
			ExtractMetadata:  request.Indexing.ExtractMetadata,
			GeneratePreviews: request.Indexing.GeneratePreviews,
		},
		Players: playerSettingsState{
			RememberMusicQueue:    request.Players.RememberMusicQueue,
			RememberVideoProgress: request.Players.RememberVideoProgress,
			AutoplayNextVideo:     request.Players.AutoplayNextVideo,
			ImageSlideshowSeconds: request.Players.ImageSlideshowSeconds,
		},
		Appearance: appearanceSettingsState{
			AccentColor:  request.Appearance.AccentColor,
			ReduceMotion: request.Appearance.ReduceMotion,
		},
		Language: languageSettingsState{
			Current: request.Language.Current,
		},
	}
}

func (state settingsState) toDto(availableLocales []string) SettingsDto {
	return SettingsDto{
		Library: LibrarySettingsDto{
			RuntimeRootPath:      strings.TrimSpace(config.AppConfig.EntryPoint),
			WatchedPaths:         append([]string(nil), state.Library.WatchedPaths...),
			RememberLastLocation: state.Library.RememberLastLocation,
			PrioritizeFavorites:  state.Library.PrioritizeFavorites,
		},
		Indexing: IndexingSettingsDto{
			WorkersEnabled:   config.AppConfig.EnableWorkers,
			ScanOnStartup:    state.Indexing.ScanOnStartup,
			ExtractMetadata:  state.Indexing.ExtractMetadata,
			GeneratePreviews: state.Indexing.GeneratePreviews,
		},
		Players: PlayerSettingsDto{
			RememberMusicQueue:    state.Players.RememberMusicQueue,
			RememberVideoProgress: state.Players.RememberVideoProgress,
			AutoplayNextVideo:     state.Players.AutoplayNextVideo,
			ImageSlideshowSeconds: state.Players.ImageSlideshowSeconds,
		},
		Appearance: AppearanceSettingsDto{
			AccentColor:  state.Appearance.AccentColor,
			ReduceMotion: state.Appearance.ReduceMotion,
		},
		Language: LanguageSettingsDto{
			Current:   state.Language.Current,
			Available: append([]string(nil), availableLocales...),
		},
	}
}

func sanitizePaths(paths []string, fallback []string) []string {
	uniquePaths := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))

	for _, path := range paths {
		trimmedPath := strings.TrimSpace(path)
		if trimmedPath == "" {
			continue
		}
		if _, exists := seen[trimmedPath]; exists {
			continue
		}
		seen[trimmedPath] = struct{}{}
		uniquePaths = append(uniquePaths, trimmedPath)
	}

	if len(uniquePaths) > 0 {
		return uniquePaths
	}

	return append([]string(nil), fallback...)
}

func normalizeSlideshowSeconds(value int, fallback int) int {
	if _, ok := allowedSlideshowSeconds[value]; ok {
		return value
	}
	return fallback
}

func normalizeAccentColor(value string, fallback string) string {
	if _, ok := allowedAccentColors[value]; ok {
		return value
	}
	return fallback
}

func resolveLocale(value string, availableLocales []string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		trimmedValue = strings.TrimSpace(config.AppConfig.Lang)
	}
	if trimmedValue == "" {
		trimmedValue = defaultLocale
	}

	if len(availableLocales) == 0 {
		return trimmedValue
	}

	for _, locale := range availableLocales {
		if locale == trimmedValue {
			return trimmedValue
		}
	}

	for _, locale := range availableLocales {
		if locale == defaultLocale {
			return defaultLocale
		}
	}

	return availableLocales[0]
}
