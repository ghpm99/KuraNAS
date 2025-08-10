package logger

import (
	"database/sql"
	"fmt"
	"time"
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
	s.withTransaction(func(tx *sql.Tx) error {
		log.UpdatedAt = time.Now()
		return s.Repository.UpdateLog(tx, log)
	})
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

	return s.UpdateLog(log)

}

func (s *LoggerService) withTransaction(fn func(tx *sql.Tx) error) error {
	return s.Repository.GetDbContext().ExecTx(fn)
}
