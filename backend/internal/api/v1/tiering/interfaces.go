package tiering

import (
	"time"

	tieringengine "nas-go/api/internal/worker/tiering"
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetSettingsDocument() (string, bool, error)
	UpsertSettingsDocument(document string) error
	// ListDemotionCandidates returns hot files idle since idleBefore and at
	// least minSizeBytes large, least-recently-used first.
	ListDemotionCandidates(minSizeBytes int64, idleBefore time.Time) ([]CandidateModel, error)
	// ListPromotionCandidates returns cold files used again since usedAfter.
	ListPromotionCandidates(usedAfter time.Time) ([]CandidateModel, error)
	// SetPhysicalPath records (or, with an empty path, clears) a file's
	// physical location in one transaction.
	SetPhysicalPath(fileID int, physicalPath string) error
	GetLastRun() (LastRunModel, bool, error)
	GetTierCounts() (TierCountsModel, error)
}

type ServiceInterface interface {
	GetSettings() (SettingsDto, error)
	UpdateSettings(dto SettingsDto) (SettingsDto, error)
	Status() (StatusDto, error)
	Usage() (TierUsageDto, error)
	WorkerInterface
}

// WorkerInterface is the narrow capability the worker engine consumes: it plans
// one migration pass (the promotions and demotions to perform plus the cold
// directory) and tells the scheduler when the next pass is due.
type WorkerInterface interface {
	// MigrationPlan resolves the persisted settings into the work for one pass.
	// enabled is false when the feature is off or not configured.
	MigrationPlan(now time.Time) (enabled bool, coldDir string, promotions []tieringengine.Promotion, demotions []tieringengine.Demotion, err error)
	// SetPhysicalPath is the callback the engine uses to persist each move.
	SetPhysicalPath(fileID int, physicalPath string) error
	// NextRunDue reports whether a new tier_migration should be enqueued now.
	NextRunDue(now time.Time) (bool, error)
}
