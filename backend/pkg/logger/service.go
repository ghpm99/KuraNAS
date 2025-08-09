package logger

import (
	"context"
	"database/sql"
	"time"
)

type LoggerService struct {
	Repository LoggerRepositoryInterface
}

func NewLoggerService(repo LoggerRepositoryInterface) *LoggerService {
	return &LoggerService{Repository: repo}
}

func (s *LoggerService) CreateLog(log LoggerModel, object interface{}) (LoggerModel, error) {
	log.CreatedAt = time.Now()
	log.StartTime = time.Now()
	log.UpdatedAt = time.Now()

	log.SetExtraData(LogExtraData{
		Data:  object,
		Error: "",
	})

	return s.withTransaction(func(tx *sql.Tx) (LoggerModel, error) {
		return s.Repository.CreateLog(tx, log)
	})
}

func (s *LoggerService) GetLogByID(id int) (LoggerModel, error) {
	return s.Repository.GetLogByID(id)
}

func (s *LoggerService) GetLogs(page, pageSize int) ([]LoggerModel, error) {
	return s.Repository.GetLogs(page, pageSize)
}

func (s *LoggerService) UpdateLog(log LoggerModel) (bool, error) {
	log.UpdatedAt = time.Now()
	return s.withTransactionBool(func(tx *sql.Tx) (bool, error) {
		return s.Repository.UpdateLog(tx, log)
	})
}

func (s *LoggerService) CompleteWithSuccessLog(log LoggerModel) (bool, error) {
	log.EndTime = sql.NullTime{Time: time.Now(), Valid: true}
	log.Status = LogStatusCompleted

	return s.UpdateLog(log)
}

func (s *LoggerService) CompleteWithErrorLog(log LoggerModel, err error) (bool, error) {
	log.EndTime = sql.NullTime{Time: time.Now(), Valid: true}
	log.Status = LogStatusFailed

	log.SetExtraData(LogExtraData{
		Error: err.Error(),
	})

	return s.UpdateLog(log)

}

func (s *LoggerService) withTransaction(fn func(tx *sql.Tx) error) (LoggerModel, error) {
	return s.Repository.GetDbContext().ExecTx(fn)
}

func (s *LoggerService) withTransactionBool(fn func(tx *sql.Tx) (bool, error)) (bool, error) {
	tx, err := s.Repository.GetDbContext().BeginTx(context.Background(), nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := fn(tx)
	if err != nil {
		return false, err
	}
	return result, tx.Commit()
}
