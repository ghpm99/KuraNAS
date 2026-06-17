package logger

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"time"

	"nas-go/api/pkg/applog"
)

type LoggerService struct {
	Repository LoggerRepositoryInterface
}

func NewLoggerService(repo LoggerRepositoryInterface) *LoggerService {
	return &LoggerService{Repository: repo}
}

func (s *LoggerService) CreateLog(log LoggerModel, object interface{}) (LoggerModel, error) {
	var loggerModelResult LoggerModel
	err := s.withTransaction(func(tx *sql.Tx) error {
		log.CreatedAt = time.Now()
		log.StartTime = time.Now()
		log.UpdatedAt = time.Now()

		err := log.SetExtraData(LogExtraData{
			Data:  object,
			Error: "",
		})

		if err != nil {
			return err
		}

		loggerModelResult, err = s.Repository.CreateLog(tx, log)
		return err
	})

	if err != nil {
		return LoggerModel{}, fmt.Errorf("error creating log: %w", err)
	}
	return loggerModelResult, nil
}

func (s *LoggerService) GetLogByID(id int) (LoggerModel, error) {
	return s.Repository.GetLogByID(id)
}

func (s *LoggerService) GetLogs(page, pageSize int) ([]LoggerModel, error) {
	return s.Repository.GetLogs(page, pageSize)
}

func (s *LoggerService) UpdateLog(log LoggerModel) error {
	err := s.withTransaction(func(tx *sql.Tx) error {
		log.UpdatedAt = time.Now()
		return s.Repository.UpdateLog(tx, log)
	})
	if err != nil {
		return fmt.Errorf("error updating log: %w", err)
	}
	return nil
}

func (s *LoggerService) CompleteWithSuccessLog(log LoggerModel) error {
	log.EndTime = sql.NullTime{Time: time.Now(), Valid: true}
	log.Status = LogStatusCompleted

	return s.UpdateLog(log)
}

func (s *LoggerService) CompleteWithErrorLog(log LoggerModel, err error) error {
	log.EndTime = sql.NullTime{Time: time.Now(), Valid: true}
	log.Status = LogStatusFailed

	log.SetExtraData(LogExtraData{
		Error: err.Error(),
	})

	// The DB log row above is for metrics and to surface a notification in the
	// app. Anything that needs investigation — the actual error and a stack
	// trace — goes to the forensic file log (pkg/applog → log/), so a production
	// failure is debuggable from the file without ever touching the database.
	applog.Error("operation failed",
		"operation", log.Name,
		"ip", log.IPAddress,
		"error", err.Error(),
		"stack", string(debug.Stack()),
	)

	return s.UpdateLog(log)

}

func (s *LoggerService) withTransaction(fn func(tx *sql.Tx) error) error {
	return s.Repository.GetDbContext().ExecTx(fn)
}
