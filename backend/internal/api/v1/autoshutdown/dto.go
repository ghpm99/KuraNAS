package autoshutdown

import "time"

// SettingsDto is both the response of GET /auto-shutdown/settings and the body
// of PUT /auto-shutdown/settings.
type SettingsDto struct {
	Enabled            bool   `json:"enabled"`
	Time               string `json:"time"`
	GracePeriodSeconds int    `json:"grace_period_seconds"`
}

// SuggestedTimeDto reports a shutdown time derived from the median of recorded
// SHUTDOWN events. Available is false when there are too few samples to trust.
type SuggestedTimeDto struct {
	Available  bool   `json:"available"`
	Time       string `json:"time"`
	SampleSize int    `json:"sample_size"`
}

// StatusDto is the response of GET /auto-shutdown/status. ShutdownAt and
// SecondsRemaining are null unless State is "warning".
type StatusDto struct {
	State            string     `json:"state"`
	ShutdownAt       *time.Time `json:"shutdown_at"`
	SecondsRemaining *int       `json:"seconds_remaining"`
	PostponedReason  string     `json:"postponed_reason"`
}

func (s settingsState) toDto() SettingsDto {
	return SettingsDto{
		Enabled:            s.Enabled,
		Time:               s.Time,
		GracePeriodSeconds: s.GracePeriodSeconds,
	}
}

func (d SettingsDto) toState() settingsState {
	return settingsState{
		Enabled:            d.Enabled,
		Time:               d.Time,
		GracePeriodSeconds: d.GracePeriodSeconds,
	}
}
