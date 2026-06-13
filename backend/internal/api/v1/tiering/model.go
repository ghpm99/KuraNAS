package tiering

import (
	"encoding/json"
	"time"
)

const (
	defaultMinAgeDays    = 90
	defaultMinSizeBytes  = 1 << 20 // 1 MiB — not worth tiering tiny files
	defaultIntervalHours = 24
	defaultBatchLimit    = 500
)

// SettingsModel is the persisted tiering configuration (one JSON document under
// the tiering_settings key in app_settings).
type SettingsModel struct {
	Enabled bool `json:"enabled"`
	// ColdDirPath is the directory on the cold volume that holds the migrated
	// bytes. It must live outside every indexed root.
	ColdDirPath string `json:"cold_dir_path"`
	// MinAgeDays demotes files untouched for at least this many days.
	MinAgeDays int `json:"min_age_days"`
	// MinSizeBytes skips files smaller than this — the I/O is not worth it.
	MinSizeBytes int64 `json:"min_size_bytes"`
	// IntervalHours is how often the migration job runs.
	IntervalHours int `json:"interval_hours"`
}

// withDefaults backfills zero values so a document persisted before a field
// existed (or hand-edited) still resolves to safe behavior.
func (m SettingsModel) withDefaults() SettingsModel {
	if m.MinAgeDays <= 0 {
		m.MinAgeDays = defaultMinAgeDays
	}
	if m.MinSizeBytes <= 0 {
		m.MinSizeBytes = defaultMinSizeBytes
	}
	if m.IntervalHours <= 0 {
		m.IntervalHours = defaultIntervalHours
	}
	return m
}

func defaultSettings() SettingsModel {
	return SettingsModel{}.withDefaults()
}

func decodeSettings(document string) (SettingsModel, error) {
	var settings SettingsModel
	if err := json.Unmarshal([]byte(document), &settings); err != nil {
		return SettingsModel{}, err
	}
	return settings.withDefaults(), nil
}

func encodeSettings(settings SettingsModel) (string, error) {
	payload, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// CandidateModel is one file the migration job will move; PhysicalPath is empty
// for a hot file selected for demotion and set for a cold file to promote.
type CandidateModel struct {
	FileID       int
	LogicalPath  string
	PhysicalPath string
	Size         int64
}

// LastRunModel mirrors the latest tier_migration row in worker_job.
type LastRunModel struct {
	JobID     int
	Status    string
	CreatedAt time.Time
	StartedAt *time.Time
	EndedAt   *time.Time
	LastError string
}

// TierCountsModel is the hot/cold split used by the analytics endpoint.
type TierCountsModel struct {
	HotFiles  int   `json:"hot_files"`
	HotBytes  int64 `json:"hot_bytes"`
	ColdFiles int   `json:"cold_files"`
	ColdBytes int64 `json:"cold_bytes"`
}
