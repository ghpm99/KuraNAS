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
	document, isFound, err := r.getDocument(queries.GetSettingsQuery)
	if err != nil {
		return "", false, fmt.Errorf("GetSettingsDocument: %w", err)
	}
	return document, isFound, nil
}

func (r *Repository) UpsertSettingsDocument(document string) error {
	if err := r.upsertDocument(queries.UpsertSettingsQuery, document); err != nil {
		return fmt.Errorf("UpsertSettingsDocument: %w", err)
	}
	return nil
}

func (r *Repository) GetCloudSettingsDocument() (string, bool, error) {
	document, isFound, err := r.getDocument(queries.GetCloudSettingsQuery)
	if err != nil {
		return "", false, fmt.Errorf("GetCloudSettingsDocument: %w", err)
	}
	return document, isFound, nil
}

func (r *Repository) UpsertCloudSettingsDocument(document string) error {
	if err := r.upsertDocument(queries.UpsertCloudSettingsQuery, document); err != nil {
		return fmt.Errorf("UpsertCloudSettingsDocument: %w", err)
	}
	return nil
}

func (r *Repository) GetRoomNamesDocument() (string, bool, error) {
	document, isFound, err := r.getDocument(queries.GetRoomNamesQuery)
	if err != nil {
		return "", false, fmt.Errorf("GetRoomNamesDocument: %w", err)
	}
	return document, isFound, nil
}

func (r *Repository) UpsertRoomNamesDocument(document string) error {
	if err := r.upsertDocument(queries.UpsertRoomNamesQuery, document); err != nil {
		return fmt.Errorf("UpsertRoomNamesDocument: %w", err)
	}
	return nil
}

func (r *Repository) getDocument(query string) (string, bool, error) {
	var document string
	isFound := true

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanErr := tx.QueryRow(query).Scan(&document)
		if errors.Is(scanErr, sql.ErrNoRows) {
			isFound = false
			return nil
		}
		return scanErr
	})
	if err != nil {
		return "", false, err
	}
	return document, isFound, nil
}

func (r *Repository) upsertDocument(query string, document string) error {
	return r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(query, document)
		return execErr
	})
}
