package backup

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/backup"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(db *database.DbContext) *Repository {
	return &Repository{DbContext: db}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) GetSettingsDocument() (string, bool, error) {
	var document string
	found := true

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanErr := tx.QueryRow(queries.GetBackupSettingsQuery).Scan(&document)
		if errors.Is(scanErr, sql.ErrNoRows) {
			found = false
			return nil
		}
		return scanErr
	})

	if err != nil {
		return "", false, fmt.Errorf("GetSettingsDocument: %w", err)
	}
	return document, found, nil
}

func (r *Repository) UpsertSettingsDocument(document string) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(queries.UpsertBackupSettingsQuery, document)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("UpsertSettingsDocument: %w", err)
	}
	return nil
}

func (r *Repository) CountPendingFiles() (int, error) {
	var pending int

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.CountPendingBackupQuery).Scan(&pending)
	})

	if err != nil {
		return 0, fmt.Errorf("CountPendingFiles: %w", err)
	}
	return pending, nil
}

func (r *Repository) GetLastRun() (LastRunModel, bool, error) {
	var run LastRunModel
	found := true

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var startedAt, endedAt sql.NullTime
		scanErr := tx.QueryRow(queries.GetLastBackupJobQuery).Scan(
			&run.JobID, &run.Status, &run.CreatedAt, &startedAt, &endedAt, &run.LastError,
		)
		if errors.Is(scanErr, sql.ErrNoRows) {
			found = false
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		if startedAt.Valid {
			run.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			run.EndedAt = &endedAt.Time
		}
		return nil
	})

	if err != nil {
		return LastRunModel{}, false, fmt.Errorf("GetLastRun: %w", err)
	}
	return run, found, nil
}

func (r *Repository) StampLastBackup(path string, at time.Time) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(queries.UpdateLastBackupByPathQuery, path, at)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("StampLastBackup: %w", err)
	}
	return nil
}
