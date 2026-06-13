package tiering

import "time"

// SettingsDto is both the response of GET /tiering/settings and the body of
// PUT /tiering/settings.
type SettingsDto struct {
	Enabled       bool   `json:"enabled"`
	ColdDirPath   string `json:"cold_dir_path"`
	MinAgeDays    int    `json:"min_age_days"`
	MinSizeBytes  int64  `json:"min_size_bytes"`
	IntervalHours int    `json:"interval_hours"`
}

// StatusDto reports the latest migration run (GET /tiering/status).
type StatusDto struct {
	Enabled   bool       `json:"enabled"`
	HasRun    bool       `json:"has_run"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	LastError string     `json:"last_error"`
}

// TierUsageDto reports the hot/cold split (GET /tiering/usage), one small
// payload per concern per the endpoint-granularity rule.
type TierUsageDto struct {
	HotFiles  int   `json:"hot_files"`
	HotBytes  int64 `json:"hot_bytes"`
	ColdFiles int   `json:"cold_files"`
	ColdBytes int64 `json:"cold_bytes"`
}

func (m SettingsModel) toDto() SettingsDto {
	return SettingsDto{
		Enabled:       m.Enabled,
		ColdDirPath:   m.ColdDirPath,
		MinAgeDays:    m.MinAgeDays,
		MinSizeBytes:  m.MinSizeBytes,
		IntervalHours: m.IntervalHours,
	}
}

func (d SettingsDto) toModel() SettingsModel {
	return SettingsModel{
		Enabled:       d.Enabled,
		ColdDirPath:   d.ColdDirPath,
		MinAgeDays:    d.MinAgeDays,
		MinSizeBytes:  d.MinSizeBytes,
		IntervalHours: d.IntervalHours,
	}.withDefaults()
}

func (m TierCountsModel) toDto() TierUsageDto {
	return TierUsageDto{
		HotFiles:  m.HotFiles,
		HotBytes:  m.HotBytes,
		ColdFiles: m.ColdFiles,
		ColdBytes: m.ColdBytes,
	}
}
