package queries

import (
	_ "embed"
)

//go:embed insert_file.sql
var InsertFileQuery string

//go:embed get_files.sql
var GetFilesQuery string

//go:embed update_file.sql
var UpdateFileQuery string

//go:embed get_children_count.sql
var GetChildrenCountQuery string

//go:embed delete_old_recent_files.sql
var DeleteOldRecentFilesQuery string

//go:embed delete_recent_file.sql
var DeleteRecentFileQuery string

//go:embed upsert_recent_file.sql
var UpsertRecentFileQuery string

//go:embed get_recent_files.sql
var GetRecentFilesQuery string

//go:embed get_recent_by_file_id.sql
var GetRecentByFileIDQuery string

//go:embed count_by_type.sql
var CountByTypeQuery string
