package backup

import (
	"time"

	backupengine "nas-go/api/internal/worker/backup"
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	// GetSettingsDocument returns the raw JSON document and whether it exists.
	GetSettingsDocument() (string, bool, error)
	UpsertSettingsDocument(document string) error
	CountPendingFiles() (int, error)
	// GetLastRun returns the latest backup_run job and whether one exists.
	GetLastRun() (LastRunModel, bool, error)
	StampLastBackup(path string, at time.Time) error
}

type ServiceInterface interface {
	GetSettings() (SettingsDto, error)
	UpdateSettings(dto SettingsDto) (SettingsDto, error)
	Status() (StatusDto, error)
	Pending() (PendingDto, error)
	WorkerInterface
}

// WorkerInterface is the narrow capability the worker engine consumes: it
// builds the run options for the backup_run step and tells the scheduler when
// the next run is due.
type WorkerInterface interface {
	// RunOptions resolves the persisted settings into engine options. enabled
	// is false when the feature is off or not configured.
	RunOptions() (enabled bool, opts backupengine.Options, err error)
	// NextRunDue reports whether a new backup_run should be enqueued now.
	NextRunDue(now time.Time) (bool, error)
}
