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

//go:embed total_space_used.sql
var TotalSpaceUsedQuery string

//go:embed count_by_format.sql
var CountByFormatQuery string

//go:embed top_files_by_size.sql
var TopFilesBySizeQuery string

//go:embed get_duplicate_files.sql
var GetDuplicateFilesQuery string

//go:embed upsert_image_metadata.sql
var UpsertImageMetadataQuery string

//go:embed get_image_metadata_by_id.sql
var GetImageMetadataByIDQuery string

//go:embed delete_image_metadata.sql
var DeleteImageMetadataQuery string

//go:embed get_audio_metadata_by_id.sql
var GetAudioMetadataByIDQuery string

//go:embed upsert_audio_metadata.sql
var UpsertAudioMetadataQuery string

//go:embed delete_audio_metadata.sql
var DeleteAudioMetadataQuery string

//go:embed get_video_metadata_by_id.sql
var GetVideoMetadataByIDQuery string

//go:embed upsert_video_metadata.sql
var UpsertVideoMetadataQuery string

//go:embed delete_video_metadata.sql
var DeleteVideoMetadataQuery string
