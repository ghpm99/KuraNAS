package notifications

import (
	"encoding/json"
	"time"
)

type NotificationDto struct {
	ID         int              `json:"id"`
	Type       string           `json:"type"`
	Title      string           `json:"title"`
	Message    string           `json:"message"`
	Metadata   any              `json:"metadata,omitempty"`
	IsRead     bool             `json:"is_read"`
	CreatedAt  time.Time        `json:"created_at"`
	GroupKey   string           `json:"group_key,omitempty"`
	GroupCount int              `json:"group_count"`
	IsGrouped  bool             `json:"is_grouped"`
}

type CreateNotificationDto struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Metadata any    `json:"metadata,omitempty"`
	GroupKey string `json:"group_key,omitempty"`
}

type UnreadCountDto struct {
	UnreadCount int `json:"unread_count"`
}

func toDto(model NotificationModel) NotificationDto {
	dto := NotificationDto{
		ID:         model.ID,
		Type:       model.Type,
		Title:      model.Title,
		Message:    model.Message,
		IsRead:     model.IsRead,
		CreatedAt:  model.CreatedAt,
		GroupCount: model.GroupCount,
		IsGrouped:  model.IsGrouped,
	}

	if model.GroupKey.Valid {
		dto.GroupKey = model.GroupKey.String
	}

	if model.Metadata.Valid && model.Metadata.String != "" {
		var meta any
		if err := json.Unmarshal([]byte(model.Metadata.String), &meta); err == nil {
			dto.Metadata = meta
		}
	}

	return dto
}
