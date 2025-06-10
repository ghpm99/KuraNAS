package logger

import (
	"context"
	"database/sql"
	"encoding/json"
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

	if object != nil {
		if jsonBytes, err := json.Marshal(object); err == nil {
			log.ExtraData = sql.NullString{String: string(jsonBytes), Valid: true}
		} else {
			log.ExtraData = sql.NullString{Valid: false}
		}
	} else {
		log.ExtraData = sql.NullString{Valid: false}
	}

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

func (s *LoggerService) CompleteLog(log LoggerModel) (bool, error) {
	log.EndTime = sql.NullTime{Time: time.Now(), Valid: true}
	log.Status = LogStatusCompleted

	return s.UpdateLog(log)

}

func (s *LoggerService) withTransaction(fn func(tx *sql.Tx) (LoggerModel, error)) (LoggerModel, error) {
	tx, err := s.Repository.GetDbContext().BeginTx(context.Background(), nil)
	if err != nil {
		return LoggerModel{}, err
	}
	defer tx.Rollback()
	result, err := fn(tx)
	if err != nil {
		return LoggerModel{}, err
	}
	return result, tx.Commit()
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
