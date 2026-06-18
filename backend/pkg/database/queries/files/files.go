package queries

import (
	_ "embed"
)

//go:embed insert_file.sql
var InsertFileQuery string

//go:embed get_file_by_id.sql
var GetFileByIDQuery string

//go:embed get_files_by_name_and_path.sql
var GetFilesByNameAndPathQuery string

//go:embed get_children_by_parent_path.sql
var GetChildrenByParentPathQuery string

//go:embed get_starred_children_by_parent_path.sql
var GetStarredChildrenByParentPathQuery string

//go:embed get_recent_children_by_parent_path.sql
var GetRecentChildrenByParentPathQuery string

//go:embed get_files_by_path.sql
var GetFilesByPathQuery string

//go:embed get_active_files.sql
var GetActiveFilesQuery string

//go:embed get_files_by_path_prefix.sql
var GetFilesByPathPrefixQuery string

//go:embed get_file_stat_by_path.sql
var GetFileStatByPathQuery string

//go:embed update_file.sql
var UpdateFileQuery string

//go:embed get_children_count.sql
var GetChildrenCountQuery string

//go:embed update_descendant_paths.sql
var UpdateDescendantPathsQuery string

//go:embed mark_deleted_subtree.sql
var MarkDeletedSubtreeQuery string

//go:embed restore_subtree.sql
var RestoreSubtreeQuery string

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

//go:embed total_space_used.sql
var TotalSpaceUsedQuery string

//go:embed count_by_format.sql
var CountByFormatQuery string

//go:embed top_files_by_size.sql
var TopFilesBySizeQuery string

//go:embed get_duplicate_files.sql
var GetDuplicateFilesQuery string

//go:embed delete_file_by_id.sql
var DeleteFileByIDQuery string
