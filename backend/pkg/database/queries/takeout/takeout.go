package queries

import (
	_ "embed"
)

//go:embed insert_takeout_import.sql
var InsertTakeoutImportQuery string

//go:embed mark_takeout_import_running.sql
var MarkTakeoutImportRunningQuery string

//go:embed close_takeout_import_completed.sql
var CloseTakeoutImportCompletedQuery string

//go:embed close_takeout_import_failed.sql
var CloseTakeoutImportFailedQuery string

//go:embed upsert_takeout_photo_sidecar.sql
var UpsertTakeoutPhotoSidecarQuery string

//go:embed select_takeout_photo_sidecar_by_path.sql
var SelectTakeoutPhotoSidecarByPathQuery string

//go:embed update_takeout_import_progress.sql
var UpdateTakeoutImportProgressQuery string

//go:embed select_takeout_imports_page.sql
var SelectTakeoutImportsPageQuery string
