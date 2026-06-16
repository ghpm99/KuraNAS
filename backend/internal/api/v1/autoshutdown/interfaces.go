package autoshutdown

import (
	"time"

	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetSettingsDocument() (string, bool, error)
	UpsertSettingsDocument(document string) error
	// GetShutdownTimeMedian returns the median time-of-day (seconds since
	// midnight) of recorded SHUTDOWN events and how many samples backed it.
	GetShutdownTimeMedian() (medianSeconds float64, sampleSize int, err error)
}

type ServiceInterface interface {
	GetSettings() (SettingsDto, error)
	UpdateSettings(dto SettingsDto) (SettingsDto, error)
	SuggestedTime() (SuggestedTimeDto, error)
	SchedulerInterface
}

// SchedulerInterface is the narrow capability the background scheduler consumes:
// it asks, on each tick, whether a shutdown is due at the given instant.
type SchedulerInterface interface {
	// DueNow reports whether the configured shutdown should fire now and, when
	// it should, the OS grace period (seconds) to apply. It is false whenever
	// the feature is disabled or the current minute is not the configured one.
	DueNow(now time.Time) (due bool, graceSeconds int, err error)
}
