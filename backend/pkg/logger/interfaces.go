package logger

import (
	"database/sql"
)

type LoggerRepositoryInterface interface {
	GetDbContext() *sql.DB
	CreateLog(tx *sql.Tx, log LoggerModel) (LoggerModel, error)
	GetLogByID(id int) (LoggerModel, error)
	GetLogs(page, pageSize int) ([]LoggerModel, error)
	UpdateLog(tx *sql.Tx, log LoggerModel) (bool, error)
}

type LoggerServiceInterface interface {
	CreateLog(log LoggerModel, object interface{}) (LoggerModel, error)
	GetLogByID(id int) (LoggerModel, error)
	GetLogs(page, pageSize int) ([]LoggerModel, error)
	UpdateLog(log LoggerModel) (bool, error)
	CompleteWithSuccessLog(log LoggerModel) (bool, error)
	CompleteWithErrorLog(log LoggerModel, err error) (bool, error)
}
