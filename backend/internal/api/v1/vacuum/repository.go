package vacuum

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/vacuum"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(db *database.DbContext) *Repository {
	return &Repository{DbContext: db}
}

func (r *Repository) GetSettingsDocument() (string, bool, error) {
	var document string
	isFound := true

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanErr := tx.QueryRow(queries.GetSettingsQuery).Scan(&document)
		if errors.Is(scanErr, sql.ErrNoRows) {
			isFound = false
			return nil
		}
		return scanErr
	})
	if err != nil {
		return "", false, fmt.Errorf("GetSettingsDocument: %w", err)
	}
	return document, isFound, nil
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
