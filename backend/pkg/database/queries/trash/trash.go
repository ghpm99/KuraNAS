package trash

import _ "embed"

//go:embed insert_trash_item.sql
var InsertTrashItemQuery string

//go:embed get_trash_items_sorted_by_deleted_at_asc.sql
var GetTrashItemsSortedByDeletedAtAscQuery string

//go:embed get_trash_items_sorted_by_deleted_at_desc.sql
var GetTrashItemsSortedByDeletedAtDescQuery string

//go:embed get_trash_items_sorted_by_name_asc.sql
var GetTrashItemsSortedByNameAscQuery string

//go:embed get_trash_items_sorted_by_name_desc.sql
var GetTrashItemsSortedByNameDescQuery string

//go:embed get_trash_items_sorted_by_size_asc.sql
var GetTrashItemsSortedBySizeAscQuery string

//go:embed get_trash_items_sorted_by_size_desc.sql
var GetTrashItemsSortedBySizeDescQuery string

//go:embed get_trash_item_by_id.sql
var GetTrashItemByIDQuery string

//go:embed get_expired_trash_items.sql
var GetExpiredTrashItemsQuery string

//go:embed get_all_trash_items.sql
var GetAllTrashItemsQuery string

//go:embed delete_trash_item.sql
var DeleteTrashItemQuery string

//go:embed count_overlapping_trash_items.sql
var CountOverlappingTrashItemsQuery string

//go:embed get_trash_summary.sql
var GetTrashSummaryQuery string

//go:embed get_retention_days.sql
var GetRetentionDaysQuery string

//go:embed upsert_retention_days.sql
var UpsertRetentionDaysQuery string
