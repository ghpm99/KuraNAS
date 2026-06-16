package autoshutdown

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
