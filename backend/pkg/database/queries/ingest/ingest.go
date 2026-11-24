package queries

import (
	_ "embed"
)

//go:embed insert_ingest_download.sql
var InsertIngestDownloadQuery string

//go:embed mark_ingest_download_running.sql
var MarkIngestDownloadRunningQuery string

//go:embed update_ingest_download_progress.sql
var UpdateIngestDownloadProgressQuery string

//go:embed update_ingest_download_title.sql
var UpdateIngestDownloadTitleQuery string

//go:embed mark_ingest_download_completed.sql
var MarkIngestDownloadCompletedQuery string

//go:embed mark_ingest_download_failed.sql
var MarkIngestDownloadFailedQuery string

//go:embed mark_ingest_download_canceled.sql
var MarkIngestDownloadCanceledQuery string
