package watchfolders

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/watch_folders"
	"time"
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

func (r *Repository) GetAll() ([]WatchFolderModel, error) {
	models := make([]WatchFolderModel, 0)

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetWatchFoldersQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var model WatchFolderModel
			if err := rows.Scan(
				&model.ID,
				&model.Path,
				&model.Label,
				&model.Enabled,
				&model.LastScanAt,
				&model.CreatedAt,
				&model.UpdatedAt,
			); err != nil {
				return err
			}
			models = append(models, model)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}

	return models, nil
}

func (r *Repository) GetByID(id int) (WatchFolderModel, error) {
	var model WatchFolderModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetWatchFolderByIDQuery, id).Scan(
			&model.ID,
			&model.Path,
			&model.Label,
			&model.Enabled,
			&model.LastScanAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
	})
	if err != nil {
		return model, fmt.Errorf("GetByID: %w", err)
	}

	return model, nil
}

func (r *Repository) Create(tx *sql.Tx, model WatchFolderModel) (WatchFolderModel, error) {
	var result WatchFolderModel
	err := tx.QueryRow(
		queries.CreateWatchFolderQuery,
		model.Path,
		model.Label,
		model.Enabled,
		model.LastScanAt,
	).Scan(
		&result.ID,
		&result.Path,
		&result.Label,
		&result.Enabled,
		&result.LastScanAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return result, fmt.Errorf("Create: %w", err)
	}

	return result, nil
}

func (r *Repository) Update(tx *sql.Tx, model WatchFolderModel) (WatchFolderModel, error) {
	var result WatchFolderModel
	err := tx.QueryRow(
		queries.UpdateWatchFolderQuery,
		model.ID,
		model.Path,
		model.Label,
		model.Enabled,
	).Scan(
		&result.ID,
		&result.Path,
		&result.Label,
		&result.Enabled,
		&result.LastScanAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return result, fmt.Errorf("Update: %w", err)
	}

	return result, nil
}

func (r *Repository) Delete(tx *sql.Tx, id int) error {
	result, err := tx.Exec(queries.DeleteWatchFolderQuery, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) UpdateLastScan(tx *sql.Tx, id int, lastScanAt time.Time) error {
	result, err := tx.Exec(queries.UpdateWatchFolderLastScanQuery, id, lastScanAt)
	if err != nil {
		return fmt.Errorf("UpdateLastScan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateLastScan rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
