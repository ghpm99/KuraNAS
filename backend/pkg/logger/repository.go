package logger

import (
	"database/sql"
	"fmt"
	queries "nas-go/api/pkg/database/queries/log"
)

type LoggerRepository struct {
	DbContext *sql.DB
}

func NewLoggerRepository(db *sql.DB) *LoggerRepository {
	return &LoggerRepository{db}
}

func (r *LoggerRepository) GetDbContext() *sql.DB {
	return r.DbContext
}

func (r *LoggerRepository) CreateLog(tx *sql.Tx, log LoggerModel) (LoggerModel, error) {
	query := queries.InsertLogQuery
	args := []any{
		log.Name, log.Description, log.Level, log.IPAddress, log.StartTime, log.EndTime,
		log.Status, log.ExtraData,
	}
	result, err := tx.Exec(query, args...)
	if err != nil {
		return log, fmt.Errorf("CreateLog: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return log, fmt.Errorf("CreateLog: %v", err)
	}
	log.ID = int(id)
	return log, nil
}

func (r *LoggerRepository) GetLogByID(id int) (LoggerModel, error) {
	row := r.DbContext.QueryRow(queries.GetLogByIDQuery, id)
	var log LoggerModel
	err := row.Scan(
		&log.ID, &log.Name, &log.Description, &log.Level, &log.IPAddress,
		&log.StartTime, &log.EndTime, &log.CreatedAt, &log.UpdatedAt,
		&log.DeletedAt, &log.Status, &log.ExtraData,
	)
	return log, err
}

func (r *LoggerRepository) GetLogs(page, pageSize int) ([]LoggerModel, error) {
	offset := (page - 1) * pageSize
	rows, err := r.DbContext.Query(queries.GetLogsQuery, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []LoggerModel
	for rows.Next() {
		var log LoggerModel
		err := rows.Scan(
			&log.ID, &log.Name, &log.Description, &log.Level, &log.IPAddress,
			&log.StartTime, &log.EndTime, &log.CreatedAt, &log.UpdatedAt,
			&log.DeletedAt, &log.Status, &log.ExtraData,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *LoggerRepository) UpdateLog(tx *sql.Tx, log LoggerModel) (bool, error) {
	query := queries.UpdateLogQuery
	args := []any{
		log.Name, log.Description, log.Level, log.IPAddress, log.StartTime, log.EndTime,
		log.Status, log.ExtraData, log.ID,
	}
	result, err := tx.Exec(query, args...)
	if err != nil {
		return false, fmt.Errorf("UpdateLog: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}
