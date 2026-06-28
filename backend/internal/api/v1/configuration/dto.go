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

// EnvFieldDto is one editable .env variable as seen by the infra wizard. For
// secret fields the actual value is never sent to the client (write-only): only
// Configured tells whether a value is currently set. For non-secret fields Value
// carries the current value (or its effective default when absent from the file).
type EnvFieldDto struct {
	Key        string `json:"key"`
	Group      string `json:"group"`
	Kind       string `json:"kind"`
	Value      string `json:"value"`
	Configured bool   `json:"configured"`
	Dangerous  bool   `json:"dangerous"`
}

// EnvConfigDto is the whole .env, grouped field list plus whether a restart is
// pending because a write happened after boot (the process still runs the old
// snapshot until restarted).
type EnvConfigDto struct {
	Fields          []EnvFieldDto `json:"fields"`
	RestartRequired bool          `json:"restart_required"`
}

// UpdateEnvConfigRequest carries only the keys the user changed. Secret keys are
// present only when the user typed a new value (empty/absent means keep current).
// Confirmed must be true when any dangerous key (DB_*, ALLOWED_ORIGINS,
// EMAIL_TOKEN_KEY) is among the changes.
type UpdateEnvConfigRequest struct {
	Changes   map[string]string `json:"changes"`
	Confirmed bool              `json:"confirmed"`
}

// TestDatabaseRequest validates a candidate database connection before it is
// persisted. An empty Password reuses the currently stored DB_PASSWORD.
type TestDatabaseRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type TestPathRequest struct {
	Path string `json:"path"`
}

// EnvTestResultDto is the outcome of a side-effecting validator (test-db /
// test-path). Message is already translated server-side.
type EnvTestResultDto struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
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
