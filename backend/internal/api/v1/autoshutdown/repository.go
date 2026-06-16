package autoshutdown

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/autoshutdown"
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
		scanErr := tx.QueryRow(queries.GetSettingsQuery).Scan(&document)
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
		_, execErr := tx.Exec(queries.UpsertSettingsQuery, document)
		return execErr
	})
	if err != nil {
		return fmt.Errorf("UpsertSettingsDocument: %w", err)
	}
	return nil
}

func (r *Repository) GetShutdownTimeMedian() (float64, int, error) {
	var medianSeconds sql.NullFloat64
	var sampleSize int

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetShutdownTimeMedianQuery).Scan(&medianSeconds, &sampleSize)
	})
	if err != nil {
		return 0, 0, fmt.Errorf("GetShutdownTimeMedian: %w", err)
	}

	if !medianSeconds.Valid {
		return 0, sampleSize, nil
	}
	return medianSeconds.Float64, sampleSize, nil
}
