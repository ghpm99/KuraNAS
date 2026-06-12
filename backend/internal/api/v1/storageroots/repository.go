package storageroots

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/storageroots"
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

func scanStorageRoot(scanner interface{ Scan(...any) error }) (StorageRootModel, error) {
	var model StorageRootModel
	err := scanner.Scan(
		&model.ID,
		&model.Path,
		&model.Label,
		&model.Enabled,
		&model.CreatedAt,
	)
	return model, err
}

func (r *Repository) GetAll() ([]StorageRootModel, error) {
	// Runs at boot before anything else; a context without a real database
	// (tests) must fail soft, not panic.
	if r.DbContext == nil || r.DbContext.GetDatabase() == nil {
		return nil, sql.ErrConnDone
	}

	models := make([]StorageRootModel, 0)

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetStorageRootsQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			model, scanErr := scanStorageRoot(rows)
			if scanErr != nil {
				return scanErr
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

func (r *Repository) GetByID(id int) (StorageRootModel, bool, error) {
	var model StorageRootModel
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanned, scanErr := scanStorageRoot(tx.QueryRow(queries.GetStorageRootByIDQuery, id))
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		model = scanned
		found = true
		return nil
	})
	if err != nil {
		return StorageRootModel{}, false, fmt.Errorf("GetByID: %w", err)
	}

	return model, found, nil
}

func (r *Repository) Create(tx *sql.Tx, model StorageRootModel) (StorageRootModel, error) {
	created, err := scanStorageRoot(tx.QueryRow(
		queries.InsertStorageRootQuery,
		model.Path,
		model.Label,
		model.Enabled,
	))
	if err != nil {
		return StorageRootModel{}, fmt.Errorf("Create: %w", err)
	}
	return created, nil
}

func (r *Repository) Update(tx *sql.Tx, model StorageRootModel) (StorageRootModel, error) {
	updated, err := scanStorageRoot(tx.QueryRow(
		queries.UpdateStorageRootQuery,
		model.ID,
		model.Label,
		model.Enabled,
	))
	if err == sql.ErrNoRows {
		return StorageRootModel{}, ErrRootNotFound
	}
	if err != nil {
		return StorageRootModel{}, fmt.Errorf("Update: %w", err)
	}
	return updated, nil
}

func (r *Repository) Delete(tx *sql.Tx, id int) error {
	result, err := tx.Exec(queries.DeleteStorageRootQuery, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete rows affected: %w", err)
	}
	if affected == 0 {
		return ErrRootNotFound
	}
	return nil
}
