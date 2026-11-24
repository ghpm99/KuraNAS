package queries

import (
	_ "embed"
)

//go:embed insert_ingest_download.sql
var InsertIngestDownloadQuery string
