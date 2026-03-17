package notifications

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/notifications"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{database}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) CreateNotification(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
	err := tx.QueryRow(
		queries.InsertNotificationQuery,
		model.Type,
		model.Title,
		model.Message,
		model.Metadata,
		model.IsRead,
		model.GroupKey,
		model.GroupCount,
		model.IsGrouped,
	).Scan(
		&model.ID,
		&model.Type,
		&model.Title,
		&model.Message,
		&model.Metadata,
		&model.IsRead,
		&model.CreatedAt,
		&model.GroupKey,
		&model.GroupCount,
		&model.IsGrouped,
	)
	if err != nil {
		return model, fmt.Errorf("CreateNotification: %w", err)
	}

	return model, nil
}

func (r *Repository) GetNotificationByID(id int) (NotificationModel, error) {
	var model NotificationModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetNotificationByIDQuery, id).Scan(
			&model.ID,
			&model.Type,
			&model.Title,
			&model.Message,
			&model.Metadata,
			&model.IsRead,
			&model.CreatedAt,
			&model.GroupKey,
			&model.GroupCount,
			&model.IsGrouped,
		)
	})
	if err != nil {
		return model, fmt.Errorf("GetNotificationByID: %w", err)
	}

	return model, nil
}

func (r *Repository) ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationModel], error) {
	response := utils.PaginationResponse[NotificationModel]{
		Items: []NotificationModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
	}

	args := []any{
		!filter.Type.HasValue,
		filter.Type.Value,
		!filter.IsRead.HasValue,
		filter.IsRead.Value,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(queries.ListNotificationsQuery, args...)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var model NotificationModel

			if scanErr := rows.Scan(
				&model.ID,
				&model.Type,
				&model.Title,
				&model.Message,
				&model.Metadata,
				&model.IsRead,
				&model.CreatedAt,
				&model.GroupKey,
				&model.GroupCount,
				&model.IsGrouped,
			); scanErr != nil {
				return scanErr
			}

			response.Items = append(response.Items, model)
		}

		return nil
	})
	if err != nil {
		return response, fmt.Errorf("ListNotifications: %w", err)
	}

	response.UpdatePagination()
	return response, nil
}

func (r *Repository) MarkAsRead(tx *sql.Tx, id int) error {
	_, err := tx.Exec(queries.MarkAsReadQuery, id)
	if err != nil {
		return fmt.Errorf("MarkAsRead: %w", err)
	}
	return nil
}

func (r *Repository) MarkAllAsRead(tx *sql.Tx) error {
	_, err := tx.Exec(queries.MarkAllAsReadQuery)
	if err != nil {
		return fmt.Errorf("MarkAllAsRead: %w", err)
	}
	return nil
}

func (r *Repository) GetUnreadCount() (int, error) {
	var count int

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetUnreadCountQuery).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("GetUnreadCount: %w", err)
	}

	return count, nil
}

func (r *Repository) FindActiveGroup(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error) {
	var model NotificationModel

	err := tx.QueryRow(
		queries.FindActiveGroupQuery,
		groupKey,
		notifType,
		fmt.Sprintf("%d", windowSeconds),
	).Scan(
		&model.ID,
		&model.Type,
		&model.Title,
		&model.Message,
		&model.Metadata,
		&model.IsRead,
		&model.CreatedAt,
		&model.GroupKey,
		&model.GroupCount,
		&model.IsGrouped,
	)
	if err != nil {
		return model, err
	}

	return model, nil
}

func (r *Repository) UpdateGroupCount(tx *sql.Tx, id int, count int, message string) error {
	_, err := tx.Exec(queries.UpdateGroupCountQuery, count, message, id)
	if err != nil {
		return fmt.Errorf("UpdateGroupCount: %w", err)
	}
	return nil
}

func (r *Repository) DeleteOldNotifications(tx *sql.Tx) error {
	_, err := tx.Exec(queries.DeleteOldNotificationsQuery)
	if err != nil {
		return fmt.Errorf("DeleteOldNotifications: %w", err)
	}
	return nil
}
