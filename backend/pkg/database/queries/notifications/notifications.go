package queries

import (
	_ "embed"
)

//go:embed insert_notification.sql
var InsertNotificationQuery string

//go:embed get_notification_by_id.sql
var GetNotificationByIDQuery string

//go:embed list_notifications.sql
var ListNotificationsQuery string

//go:embed mark_as_read.sql
var MarkAsReadQuery string

//go:embed mark_all_as_read.sql
var MarkAllAsReadQuery string

//go:embed get_unread_count.sql
var GetUnreadCountQuery string

//go:embed find_active_group.sql
var FindActiveGroupQuery string

//go:embed update_group_count.sql
var UpdateGroupCountQuery string

//go:embed delete_old_notifications.sql
var DeleteOldNotificationsQuery string
