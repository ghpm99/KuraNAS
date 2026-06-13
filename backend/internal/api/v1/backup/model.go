package backup

import (
	"encoding/json"
	"time"
)

const (
	defaultRetentionDays = 30
	defaultIntervalHours = 24
)

// SettingsModel is the persisted backup configuration (one JSON document under
// the backup_settings key in app_settings).
type SettingsModel struct {
	Enabled         bool   `json:"enabled"`
	DestinationPath string `json:"destination_path"`
	RetentionDays   int    `json:"retention_days"`
	IntervalHours   int    `json:"interval_hours"`
}

// withDefaults backfills zero values so documents persisted before a field
// existed (or hand-edited ones) resolve to safe behavior.
func (m SettingsModel) withDefaults() SettingsModel {
	if m.RetentionDays <= 0 {
		m.RetentionDays = defaultRetentionDays
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

// LastRunModel mirrors the latest backup_run row in worker_job.
type LastRunModel struct {
	JobID     int
	Status    string
	CreatedAt time.Time
	StartedAt *time.Time
	EndedAt   *time.Time
	LastError string
}
