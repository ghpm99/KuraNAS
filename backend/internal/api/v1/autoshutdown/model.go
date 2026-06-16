package autoshutdown

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	// defaultTime is used when no document exists yet; the feature is off by
	// default so this only seeds the UI's initial value.
	defaultTime = "03:00"
	// defaultGracePeriodSeconds gives a short window to cancel/save work before
	// the machine powers off. Configurable per the user's needs.
	defaultGracePeriodSeconds = 60
	// maxGracePeriodSeconds caps the OS countdown at 24h, the documented ceiling
	// of Windows `shutdown /t`.
	maxGracePeriodSeconds = 86400
	// minSampleSize is how many recorded SHUTDOWN events the median needs before
	// the suggestion is considered meaningful.
	minSampleSize = 3
)

// settingsState is the persisted auto-shutdown configuration (one JSON document
// under the auto_shutdown_settings key in app_settings).
type settingsState struct {
	// Enabled gates the scheduler; when false nothing ever fires.
	Enabled bool `json:"enabled"`
	// Time is the local time-of-day to shut down, formatted "HH:MM" (24h).
	Time string `json:"time"`
	// GracePeriodSeconds is the OS countdown before the machine powers off.
	GracePeriodSeconds int `json:"grace_period_seconds"`
}

// withDefaults backfills zero values so a document persisted before a field
// existed (or hand-edited) still resolves to safe behavior.
func (s settingsState) withDefaults() settingsState {
	if strings.TrimSpace(s.Time) == "" {
		s.Time = defaultTime
	}
	if s.GracePeriodSeconds <= 0 {
		s.GracePeriodSeconds = defaultGracePeriodSeconds
	}
	return s
}

func defaultSettings() settingsState {
	return settingsState{}.withDefaults()
}

func decodeSettings(document string) (settingsState, error) {
	var settings settingsState
	if err := json.Unmarshal([]byte(document), &settings); err != nil {
		return settingsState{}, err
	}
	return settings.withDefaults(), nil
}

func encodeSettings(settings settingsState) (string, error) {
	payload, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// normalizeTime validates and canonicalizes an "HH:MM" 24h string. It returns
// the zero-padded form ("3:0" -> "03:00") or an error for anything out of range.
func normalizeTime(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("time must be HH:MM")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 23 {
		return "", fmt.Errorf("time hours out of range")
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil || minutes < 0 || minutes > 59 {
		return "", fmt.Errorf("time minutes out of range")
	}

	return fmt.Sprintf("%02d:%02d", hours, minutes), nil
}

// secondsToTime converts seconds-since-midnight into an "HH:MM" string, wrapping
// any value at or beyond 24h back into the day.
func secondsToTime(seconds float64) string {
	total := int(seconds+0.5) % maxGracePeriodSeconds
	if total < 0 {
		total += maxGracePeriodSeconds
	}
	return fmt.Sprintf("%02d:%02d", total/3600, (total%3600)/60)
}
