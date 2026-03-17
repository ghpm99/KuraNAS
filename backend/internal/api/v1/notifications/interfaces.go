package notifications

import (
	"database/sql"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type NotificationFilter struct {
	Type   utils.Optional[string]
	IsRead utils.Optional[bool]
}

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateNotification(tx *sql.Tx, model NotificationModel) (NotificationModel, error)
	GetNotificationByID(id int) (NotificationModel, error)
	ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationModel], error)
	MarkAsRead(tx *sql.Tx, id int) error
	MarkAllAsRead(tx *sql.Tx) error
	GetUnreadCount() (int, error)
	FindActiveGroup(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error)
	UpdateGroupCount(tx *sql.Tx, id int, count int, message string) error
	DeleteOldNotifications(tx *sql.Tx) error
}

type ServiceInterface interface {
	GetNotificationByID(id int) (NotificationDto, error)
	ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationDto], error)
	MarkAsRead(id int) error
	MarkAllAsRead() error
	GetUnreadCount() (UnreadCountDto, error)
	GroupOrCreate(dto CreateNotificationDto) (NotificationDto, error)
	CleanupOldNotifications() error
}
