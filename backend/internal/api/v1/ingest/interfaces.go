package ingest

import (
	"database/sql"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
)

// ServiceInterface is the mocking boundary the handler depends on.
type ServiceInterface interface {
	Fetch(request FetchRequestDto) (int, error)
	ListTargets() []TargetDto
	ListPresets() []PresetDto
}

// jobEnqueuer is the slice of the jobs repository the download service needs to
// enqueue a background fetch. Declared locally so the service depends on the
// narrowest contract and tests can fake it without the full repository.
type jobEnqueuer interface {
	GetDbContext() *database.DbContext
	CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error)
	CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error)
}
