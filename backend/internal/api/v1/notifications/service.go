package notifications

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"nas-go/api/pkg/utils"
)

var (
	ErrInvalidNotificationID = errors.New("invalid notification id")
	ErrNotificationNotFound  = errors.New("notification not found")
)

const groupWindowSeconds = 60

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return s.Repository.GetDbContext().ExecTx(fn)
}

func (s *Service) GetNotificationByID(id int) (NotificationDto, error) {
	if id <= 0 {
		return NotificationDto{}, ErrInvalidNotificationID
	}

	model, err := s.Repository.GetNotificationByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NotificationDto{}, ErrNotificationNotFound
		}
		return NotificationDto{}, err
	}

	return toDto(model), nil
}

func (s *Service) ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationDto], error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	modelsPagination, err := s.Repository.ListNotifications(filter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[NotificationDto]{}, err
	}

	response := utils.PaginationResponse[NotificationDto]{
		Items: make([]NotificationDto, 0, len(modelsPagination.Items)),
		Pagination: utils.Pagination{
			Page:     modelsPagination.Pagination.Page,
			PageSize: modelsPagination.Pagination.PageSize,
			HasNext:  modelsPagination.Pagination.HasNext,
			HasPrev:  modelsPagination.Pagination.HasPrev,
		},
	}

	for _, model := range modelsPagination.Items {
		response.Items = append(response.Items, toDto(model))
	}

	return response, nil
}

func (s *Service) MarkAsRead(id int) error {
	if id <= 0 {
		return ErrInvalidNotificationID
	}

	_, err := s.Repository.GetNotificationByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotificationNotFound
		}
		return err
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.MarkAsRead(tx, id)
	})
}

func (s *Service) MarkAllAsRead() error {
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.MarkAllAsRead(tx)
	})
}

func (s *Service) GetUnreadCount() (UnreadCountDto, error) {
	count, err := s.Repository.GetUnreadCount()
	if err != nil {
		return UnreadCountDto{}, err
	}
	return UnreadCountDto{UnreadCount: count}, nil
}

func (s *Service) GroupOrCreate(dto CreateNotificationDto) (NotificationDto, error) {
	var result NotificationDto

	err := s.withTransaction(func(tx *sql.Tx) error {
		// Errors are never grouped
		if dto.GroupKey == "" || dto.Type == string(NotificationTypeError) {
			model, createErr := s.createNotification(tx, dto, false)
			if createErr != nil {
				return createErr
			}
			result = toDto(model)
			return nil
		}

		// Try to find an active group
		existing, findErr := s.Repository.FindActiveGroup(tx, dto.GroupKey, dto.Type, groupWindowSeconds)
		if findErr != nil {
			if !errors.Is(findErr, sql.ErrNoRows) {
				return findErr
			}

			// No active group found — create new grouped notification
			model, createErr := s.createNotification(tx, dto, true)
			if createErr != nil {
				return createErr
			}
			result = toDto(model)
			return nil
		}

		// Active group found — increment count and update message
		newCount := existing.GroupCount + 1
		newMessage := fmt.Sprintf("%d %s", newCount, dto.Title)

		if err := s.Repository.UpdateGroupCount(tx, existing.ID, newCount, newMessage); err != nil {
			return err
		}

		existing.GroupCount = newCount
		result = toDto(existing)
		result.Message = newMessage
		return nil
	})

	if err != nil {
		return NotificationDto{}, fmt.Errorf("GroupOrCreate: %w", err)
	}

	return result, nil
}

func (s *Service) CleanupOldNotifications() error {
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.DeleteOldNotifications(tx)
	})
}

func (s *Service) createNotification(tx *sql.Tx, dto CreateNotificationDto, isGrouped bool) (NotificationModel, error) {
	model := NotificationModel{
		Type:       dto.Type,
		Title:      dto.Title,
		Message:    dto.Message,
		IsRead:     false,
		GroupCount: 1,
		IsGrouped:  isGrouped,
	}

	if dto.GroupKey != "" {
		model.GroupKey = sql.NullString{String: dto.GroupKey, Valid: true}
	}

	if dto.Metadata != nil {
		metaBytes, err := json.Marshal(dto.Metadata)
		if err != nil {
			return NotificationModel{}, fmt.Errorf("marshal metadata: %w", err)
		}
		model.Metadata = sql.NullString{String: string(metaBytes), Valid: true}
	}

	return s.Repository.CreateNotification(tx, model)
}
