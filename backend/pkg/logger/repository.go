package logger

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/log"
)

type LoggerRepository struct {
	DbContext *database.DbContext
}

func NewLoggerRepository(db *database.DbContext) *LoggerRepository {
	return &LoggerRepository{db}
}

func (r *LoggerRepository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *LoggerRepository) CreateLog(tx *sql.Tx, log LoggerModel) (LoggerModel, error) {

	query := queries.InsertLogQuery

	args := []any{
		log.Name, log.Description, log.Level, log.IPAddress, log.StartTime, log.EndTime,
		log.Status, log.ExtraData,
	}

	var id int
	err := tx.QueryRow(query, args...).Scan(&id)

	if err != nil {
		return log, fmt.Errorf("CreateLog: %v", err)
	}

	log.ID = id
	return log, nil
}

func (r *LoggerRepository) GetLogByID(id int) (LoggerModel, error) {
	var log LoggerModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetLogByIDQuery, id)

		err := row.Scan(
			&log.ID, &log.Name, &log.Description, &log.Level, &log.IPAddress,
			&log.StartTime, &log.EndTime, &log.CreatedAt, &log.UpdatedAt,
			&log.DeletedAt, &log.Status, &log.ExtraData,
		)
		return err
	})

	if err != nil {
		return LoggerModel{}, fmt.Errorf("falha ao obter log por ID: %w", err)
	}

	return log, nil
}

func (r *LoggerRepository) GetLogs(page, pageSize int) ([]LoggerModel, error) {
	var logs []LoggerModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		offset := (page - 1) * pageSize
		rows, err := tx.Query(queries.GetLogsQuery, pageSize, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var log LoggerModel
			err := rows.Scan(
				&log.ID, &log.Name, &log.Description, &log.Level, &log.IPAddress,
				&log.StartTime, &log.EndTime, &log.CreatedAt, &log.UpdatedAt,
				&log.DeletedAt, &log.Status, &log.ExtraData,
			)
			if err != nil {
				return err
			}
			logs = append(logs, log)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao obter logs: %w", err)
	}

	return logs, nil
}

func (r *LoggerRepository) UpdateLog(tx *sql.Tx, log LoggerModel) error {
	query := queries.UpdateLogQuery
	args := []any{
		log.Name, log.Description, log.Level, log.IPAddress, log.StartTime, log.EndTime,
		log.Status, log.ExtraData, log.ID,
	}
	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("UpdateLog: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 1 {
		return fmt.Errorf("multiples rows affected")
	}
	return nil
}
