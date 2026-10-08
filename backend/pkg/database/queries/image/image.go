package queries

import (
	_ "embed"
)

//go:embed upsert_image_metadata.sql
var UpsertImageMetadataQuery string

//go:embed get_image_metadata_by_id.sql
var GetImageMetadataByIDQuery string

//go:embed delete_image_metadata.sql
var DeleteImageMetadataQuery string

//go:embed get_images.sql
var GetImagesQuery string

//go:embed count_pending_ai_classification.sql
var CountPendingAIClassificationQuery string

//go:embed select_pending_ai_classification.sql
var SelectPendingAIClassificationQuery string

//go:embed get_image_summary_by_file_id.sql
var GetImageSummaryByFileIDQuery string

//go:embed library_list_select.sql
var LibraryListSelectQuery string

//go:embed library_count_select.sql
var LibraryCountSelectQuery string

//go:embed library_timeline_select.sql
var LibraryTimelineSelectQuery string

//go:embed library_scope.sql
var LibraryScopeQuery string

//go:embed library_filter_name.sql
var LibraryFilterNameQuery string

//go:embed library_filter_category.sql
var LibraryFilterCategoryQuery string

//go:embed library_filter_starred.sql
var LibraryFilterStarredQuery string

//go:embed library_filter_format.sql
var LibraryFilterFormatQuery string

//go:embed library_filter_taken_from.sql
var LibraryFilterTakenFromQuery string

//go:embed library_filter_taken_to.sql
var LibraryFilterTakenToQuery string

//go:embed library_filter_camera.sql
var LibraryFilterCameraQuery string

//go:embed library_filter_folder.sql
var LibraryFilterFolderQuery string

//go:embed library_filter_dated_only.sql
var LibraryFilterDatedOnlyQuery string

//go:embed library_seek_taken_before.sql
var LibrarySeekTakenBeforeQuery string

//go:embed library_keyset_after.sql
var LibraryKeysetAfterQuery string

//go:embed library_order_taken_at_desc.sql
var LibraryOrderTakenAtDescQuery string

//go:embed library_order_taken_at_asc.sql
var LibraryOrderTakenAtAscQuery string

//go:embed library_order_name_asc.sql
var LibraryOrderNameAscQuery string

//go:embed library_order_name_desc.sql
var LibraryOrderNameDescQuery string

//go:embed library_order_size_asc.sql
var LibraryOrderSizeAscQuery string

//go:embed library_order_size_desc.sql
var LibraryOrderSizeDescQuery string

//go:embed library_limit.sql
var LibraryLimitQuery string

//go:embed library_limit_offset.sql
var LibraryLimitOffsetQuery string

//go:embed library_timeline_group.sql
var LibraryTimelineGroupQuery string

//go:embed library_folders.sql
var LibraryFoldersQuery string
