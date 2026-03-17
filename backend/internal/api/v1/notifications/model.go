package notifications

import (
	"database/sql"
	"time"
)

type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSystem  NotificationType = "system"
)

type NotificationModel struct {
	ID         int
	Type       string
	Title      string
	Message    string
	Metadata   sql.NullString
	IsRead     bool
	CreatedAt  time.Time
	GroupKey   sql.NullString
	GroupCount int
	IsGrouped  bool
}
