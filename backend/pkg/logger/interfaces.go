package logger

import (
	"database/sql"
	"nas-go/api/pkg/database"
)

type LoggerRepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateLog(tx *sql.Tx, log LoggerModel) (LoggerModel, error)
	GetLogByID(id int) (LoggerModel, error)
	GetLogs(page, pageSize int) ([]LoggerModel, error)
	UpdateLog(tx *sql.Tx, log LoggerModel) error
}

type LoggerServiceInterface interface {
	CreateLog(log LoggerModel, object interface{}) (LoggerModel, error)
	GetLogByID(id int) (LoggerModel, error)
	GetLogs(page, pageSize int) ([]LoggerModel, error)
	UpdateLog(log LoggerModel) error
	CompleteWithSuccessLog(log LoggerModel) error
	CompleteWithErrorLog(log LoggerModel, err error) error
}
