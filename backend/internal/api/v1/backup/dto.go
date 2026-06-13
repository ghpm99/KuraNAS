package backup

import "time"

// SettingsDto is both the response of GET /backup/settings and the body of
// PUT /backup/settings.
type SettingsDto struct {
	Enabled         bool   `json:"enabled"`
	DestinationPath string `json:"destination_path"`
	RetentionDays   int    `json:"retention_days"`
	IntervalHours   int    `json:"interval_hours"`
}

// StatusDto reports the latest backup run (GET /backup/status).
type StatusDto struct {
	Enabled   bool       `json:"enabled"`
	HasRun    bool       `json:"has_run"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	LastError string     `json:"last_error"`
}

// PendingDto counts files not yet (or no longer) covered by a backup
// (GET /backup/pending).
type PendingDto struct {
	PendingFiles int `json:"pending_files"`
}

func (m SettingsModel) toDto() SettingsDto {
	return SettingsDto{
		Enabled:         m.Enabled,
		DestinationPath: m.DestinationPath,
		RetentionDays:   m.RetentionDays,
		IntervalHours:   m.IntervalHours,
	}
}

func (d SettingsDto) toModel() SettingsModel {
	return SettingsModel{
		Enabled:         d.Enabled,
		DestinationPath: d.DestinationPath,
		RetentionDays:   d.RetentionDays,
		IntervalHours:   d.IntervalHours,
	}.withDefaults()
}
