package configuration

// settingsState is the persisted shape of system preferences (stored as one
// JSON document under the system_preferences configuration key).
type settingsState struct {
	Library    librarySettingsState    `json:"library"`
	Indexing   indexingSettingsState   `json:"indexing"`
	AI         aiSettingsState         `json:"ai"`
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

// aiSettingsState stores AI toggles as pointers so an absent field in a document
// persisted before the section existed resolves to the safe default (enabled),
// instead of the bool zero value (disabled).
type aiSettingsState struct {
	ImageClassification *bool `json:"image_classification,omitempty"`
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
