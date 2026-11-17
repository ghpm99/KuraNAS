package configuration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"nas-go/api/pkg/applog"
	"time"
)

// IndexingPreferences is the indexing behavior the user chose in Settings, as
// the worker pipeline and the periodic scan scheduler consume it.
type IndexingPreferences struct {
	ExtractMetadata   bool
	GeneratePreviews  bool
	IndexDocumentText bool
	ScanInterval      time.Duration
}

// IndexingPreferencesReader gives the indexing pipeline the current preferences.
// Each call reflects the latest saved settings, so a change applies immediately.
type IndexingPreferencesReader interface {
	IndexingPreferences() IndexingPreferences
}

// DefaultIndexingPreferences returns the preferences used when nothing is saved
// or the saved document cannot be read: every stage on, scan every 24 hours.
func DefaultIndexingPreferences() IndexingPreferences {
	return IndexingPreferences{
		ExtractMetadata:   true,
		GeneratePreviews:  true,
		IndexDocumentText: defaultIndexDocumentText,
		ScanInterval:      scanIntervalFromHours(defaultScanIntervalHours),
	}
}

// IndexingPreferences reads the stored settings document on every call (one
// indexed lookup, no locale listing, no cache, so the PUT takes effect at once).
// On a read failure it logs and returns the defaults.
func (s *Service) IndexingPreferences() IndexingPreferences {
	payload, err := s.Repository.GetSettingsDocument(settingsStorageKey)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultIndexingPreferences()
	}
	if err != nil {
		applog.Error("configuration: load indexing preferences failed, using defaults", "error", err)
		return DefaultIndexingPreferences()
	}

	var stored settingsState
	if err := json.Unmarshal([]byte(payload), &stored); err != nil {
		applog.Error("configuration: decode indexing preferences failed, using defaults", "error", err)
		return DefaultIndexingPreferences()
	}

	return indexingPreferencesFromState(stored.Indexing)
}

func indexingPreferencesFromState(indexing indexingSettingsState) IndexingPreferences {
	return IndexingPreferences{
		ExtractMetadata:   indexing.ExtractMetadata,
		GeneratePreviews:  indexing.GeneratePreviews,
		IndexDocumentText: derefBool(indexing.IndexDocumentText, defaultIndexDocumentText),
		ScanInterval:      scanIntervalFromHours(normalizeScanIntervalHours(indexing.ScanIntervalHours, defaultScanIntervalHours)),
	}
}

func scanIntervalFromHours(hours int) time.Duration {
	return time.Duration(hours) * time.Hour
}
