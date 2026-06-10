package configuration

import (
	"nas-go/api/internal/config"
	"strings"
)

const (
	settingsStorageKey           = "system_preferences"
	defaultLocale                = "en-US"
	defaultAccentColor           = "violet"
	defaultAIImageClassification = true
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
		AI: aiSettingsState{
			ImageClassification: boolPtr(defaultAIImageClassification),
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
	normalized.AI.ImageClassification = resolveBoolPtr(candidate.AI.ImageClassification, defaults.AI.ImageClassification)
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
		AI: aiSettingsState{
			ImageClassification: boolPtr(request.AI.ImageClassification),
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
		AI: AISettingsDto{
			ImageClassification: derefBool(state.AI.ImageClassification, defaultAIImageClassification),
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

func boolPtr(value bool) *bool {
	return &value
}

func derefBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func resolveBoolPtr(candidate *bool, fallback *bool) *bool {
	if candidate != nil {
		return candidate
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
