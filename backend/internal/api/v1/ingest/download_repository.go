package ingest

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/ingest"
)

// QueuedDownload is a download request about to be recorded: the job that will
// run it and the destination as storage root plus relative subfolder.
type QueuedDownload struct {
	JobID          int
	URL            string
	Preset         string
	TargetRootID   int
	TargetRootName string
	Subfolder      string
}

type DownloadRepository struct {
	DbContext *database.DbContext
}

func NewDownloadRepository(dbContext *database.DbContext) *DownloadRepository {
	return &DownloadRepository{DbContext: dbContext}
}

// CreateQueuedDownload records a request in state queued inside the caller's
// transaction and returns the new download id.
func (r *DownloadRepository) CreateQueuedDownload(transaction *sql.Tx, download QueuedDownload) (int, error) {
	var downloadID int
	err := transaction.QueryRow(
		queries.InsertIngestDownloadQuery,
		download.JobID,
		download.URL,
		download.Preset,
		download.TargetRootID,
		download.TargetRootName,
		download.Subfolder,
	).Scan(&downloadID)
	if err != nil {
		return 0, fmt.Errorf("CreateQueuedDownload: %w", err)
	}
	return downloadID, nil
}
