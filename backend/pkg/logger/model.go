package logger

import (
	"database/sql"
	"encoding/json"
	"time"
)

type LogLevel string

const (
	LogLevelDebug    LogLevel = "DEBUG"
	LogLevelInfo     LogLevel = "INFO"
	LogLevelWarning  LogLevel = "WARNING"
	LogLevelError    LogLevel = "ERROR"
	LogLevelCritical LogLevel = "CRITICAL"
)

type LogStatus string

const (
	LogStatusPending   LogStatus = "PENDING"
	LogStatusCompleted LogStatus = "COMPLETED"
	LogStatusFailed    LogStatus = "FAILED"
)

type LogExtraData struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type LoggerModel struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Level       LogLevel       `json:"level"`
	IPAddress   string         `json:"ip_address,omitempty"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     sql.NullTime   `json:"end_time,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   sql.NullTime   `json:"deleted_at,omitempty"`
	Status      LogStatus      `json:"status"`
	ExtraData   sql.NullString `json:"extra_data,omitempty"`
}

func (loggerModel *LoggerModel) SetExtraData(extraData LogExtraData) error {

	if extraData.Data == nil && loggerModel.ExtraData.Valid {
		var existingData LogExtraData
		if err := json.Unmarshal([]byte(loggerModel.ExtraData.String), &existingData); err == nil {
			extraData.Data = existingData.Data
		}
	}

	if jsonBytes, err := json.Marshal(extraData); err == nil {
		loggerModel.ExtraData = sql.NullString{String: string(jsonBytes), Valid: true}
	} else {
		loggerModel.ExtraData = sql.NullString{Valid: false}
		return err
	}

	return nil
}
