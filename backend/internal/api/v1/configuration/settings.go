package configuration

import (
	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
	"path/filepath"
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
	return settingsState{
		Indexing: indexingSettingsState{
			ScanOnStartup:    true,
			ExtractMetadata:  true,
			GeneratePreviews: true,
		},
		Captures: capturesSettingsState{
			SavePath: defaultCapturesPath(),
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
	normalized.Indexing.ScanOnStartup = candidate.Indexing.ScanOnStartup
	normalized.Indexing.ExtractMetadata = candidate.Indexing.ExtractMetadata
	normalized.Indexing.GeneratePreviews = candidate.Indexing.GeneratePreviews
	normalized.Captures.SavePath = sanitizeCapturePath(candidate.Captures.SavePath, defaults.Captures.SavePath)
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
		Indexing: indexingSettingsState{
			ScanOnStartup:    request.Indexing.ScanOnStartup,
			ExtractMetadata:  request.Indexing.ExtractMetadata,
			GeneratePreviews: request.Indexing.GeneratePreviews,
		},
		Captures: capturesSettingsState{
			SavePath: request.Captures.SavePath,
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
		Indexing: IndexingSettingsDto{
			WorkersEnabled:   config.AppConfig.EnableWorkers,
			ScanOnStartup:    state.Indexing.ScanOnStartup,
			ExtractMetadata:  state.Indexing.ExtractMetadata,
			GeneratePreviews: state.Indexing.GeneratePreviews,
		},
		Captures: CapturesSettingsDto{
			SavePath:     state.Captures.SavePath,
			DefaultPath:  defaultCapturesPath(),
			StorageRoots: storageRootPaths(),
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

// defaultCapturesPath is the out-of-roots fallback location for captures: a
// sibling of the entry point named "kuranas-capturas". Keeping it outside every
// storage root means captures are never indexed/watched by default.
func defaultCapturesPath() string {
	entryPoint := strings.TrimSpace(config.AppConfig.EntryPoint)
	if entryPoint == "" {
		return "kuranas-capturas"
	}
	return filepath.Join(filepath.Dir(filepath.Clean(entryPoint)), "kuranas-capturas")
}

// sanitizeCapturePath cleans the stored captures path; an empty value falls back
// to the default. Containment-against-roots is enforced separately at update
// time (validateCapturesPath), so a document persisted earlier is never rejected
// here — only normalized.
func sanitizeCapturePath(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return filepath.Clean(trimmed)
}

// storageRootPaths returns the enabled storage root paths, surfaced to the client
// so the settings UI can warn that the captures path must stay out of them.
func storageRootPaths() []string {
	enabled := roots.Enabled()
	paths := make([]string, 0, len(enabled))
	for _, root := range enabled {
		paths = append(paths, root.Path)
	}
	return paths
}

// applyRuntimeCapturesPath publishes the captures path into the runtime config so
// the captures domain saves there immediately, without a restart.
func applyRuntimeCapturesPath(savePath string) {
	config.AppConfig.CapturesPath = strings.TrimSpace(savePath)
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
