package configuration

type SettingsDto struct {
	Indexing   IndexingSettingsDto   `json:"indexing"`
	Captures   CapturesSettingsDto   `json:"captures"`
	AI         AISettingsDto         `json:"ai"`
	Players    PlayerSettingsDto     `json:"players"`
	Appearance AppearanceSettingsDto `json:"appearance"`
	Language   LanguageSettingsDto   `json:"language"`
}

// CapturesSettingsDto exposes the configured capture save path plus the
// read-only default and the storage roots the UI must keep the path out of.
type CapturesSettingsDto struct {
	SavePath     string   `json:"save_path"`
	DefaultPath  string   `json:"default_path"`
	StorageRoots []string `json:"storage_roots"`
}

type IndexingSettingsDto struct {
	WorkersEnabled   bool `json:"workers_enabled"`
	ScanOnStartup    bool `json:"scan_on_startup"`
	ExtractMetadata  bool `json:"extract_metadata"`
	GeneratePreviews bool `json:"generate_previews"`
}

// AISettingsDto toggles AI usage per feature so heavy/expensive AI work can be
// disabled without removing providers. Each field gates one concrete feature.
type AISettingsDto struct {
	ImageClassification bool `json:"image_classification"`
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
	Indexing   IndexingSettingsRequest   `json:"indexing"`
	Captures   CapturesSettingsRequest   `json:"captures"`
	AI         AISettingsRequest         `json:"ai"`
	Players    PlayerSettingsRequest     `json:"players"`
	Appearance AppearanceSettingsRequest `json:"appearance"`
	Language   LanguageSettingsRequest   `json:"language"`
}

type CapturesSettingsRequest struct {
	SavePath string `json:"save_path"`
}

type IndexingSettingsRequest struct {
	ScanOnStartup    bool `json:"scan_on_startup"`
	ExtractMetadata  bool `json:"extract_metadata"`
	GeneratePreviews bool `json:"generate_previews"`
}

type AISettingsRequest struct {
	ImageClassification bool `json:"image_classification"`
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
