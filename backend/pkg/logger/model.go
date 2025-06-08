package logger

import (
	"database/sql"
	"time"
)

type LoggerModel struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Level       string         `json:"level"`
	IPAddress   string         `json:"ip_address,omitempty"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     *time.Time     `json:"end_time,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
	Status      string         `json:"status"`
	ExtraData   sql.NullString `json:"extra_data,omitempty"`
}
