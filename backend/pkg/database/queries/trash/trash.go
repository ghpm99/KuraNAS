package trash

import _ "embed"

//go:embed insert_trash_item.sql
var InsertTrashItemQuery string

//go:embed get_trash_items.sql
var GetTrashItemsQuery string

//go:embed get_trash_item_by_id.sql
var GetTrashItemByIDQuery string

//go:embed get_expired_trash_items.sql
var GetExpiredTrashItemsQuery string

//go:embed get_all_trash_items.sql
var GetAllTrashItemsQuery string

//go:embed delete_trash_item.sql
var DeleteTrashItemQuery string

//go:embed get_retention_days.sql
var GetRetentionDaysQuery string

//go:embed upsert_retention_days.sql
var UpsertRetentionDaysQuery string
