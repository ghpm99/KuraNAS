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

// DownloadProgress is what yt-dlp reported on one `[download]` line. Speed and
// ETA are nil when the tool does not know them yet.
type DownloadProgress struct {
	Percent             int
	SpeedBytesPerSecond *int64
	ETASeconds          *int
}

// DownloadRecorderInterface is what the worker needs to keep a download record
// in step with the running yt-dlp process.
type DownloadRecorderInterface interface {
	MarkRunning(downloadID int) error
	UpdateProgress(downloadID int, progress DownloadProgress) error
	UpdateTitle(downloadID int, title string) error
	MarkCompleted(downloadID int, title string, finalPath string) error
	MarkFailed(downloadID int, translatedError string) error
	MarkCanceled(downloadID int) error
}

func (r *DownloadRepository) MarkRunning(downloadID int) error {
	return r.execute("MarkRunning", queries.MarkIngestDownloadRunningQuery, downloadID)
}

func (r *DownloadRepository) UpdateProgress(downloadID int, progress DownloadProgress) error {
	return r.execute("UpdateProgress", queries.UpdateIngestDownloadProgressQuery,
		downloadID, progress.Percent, progress.SpeedBytesPerSecond, progress.ETASeconds)
}

func (r *DownloadRepository) UpdateTitle(downloadID int, title string) error {
	return r.execute("UpdateTitle", queries.UpdateIngestDownloadTitleQuery, downloadID, title)
}

// MarkCompleted ends a download that produced its file at finalPath.
func (r *DownloadRepository) MarkCompleted(downloadID int, title string, finalPath string) error {
	return r.execute("MarkCompleted", queries.MarkIngestDownloadCompletedQuery, downloadID, title, finalPath)
}

// MarkFailed ends a download with its already translated failure reason.
func (r *DownloadRepository) MarkFailed(downloadID int, translatedError string) error {
	return r.execute("MarkFailed", queries.MarkIngestDownloadFailedQuery, downloadID, translatedError)
}

func (r *DownloadRepository) MarkCanceled(downloadID int) error {
	return r.execute("MarkCanceled", queries.MarkIngestDownloadCanceledQuery, downloadID)
}

func (r *DownloadRepository) execute(operation string, query string, arguments ...any) error {
	if _, err := r.DbContext.GetDatabase().Exec(query, arguments...); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return nil
}
