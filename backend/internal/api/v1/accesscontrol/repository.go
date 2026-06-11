package accesscontrol

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/accesscontrol"
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

func scanAllowedIP(scanner interface{ Scan(...any) error }) (AllowedIPModel, error) {
	var model AllowedIPModel
	err := scanner.Scan(
		&model.ID,
		&model.CIDR,
		&model.Label,
		&model.Enabled,
		&model.CreatedAt,
	)
	return model, err
}

func (r *Repository) GetAll() ([]AllowedIPModel, error) {
	// GetAll runs at service construction (initial cache load); a context
	// built without a real database (tests) must fail soft, not panic.
	if r.DbContext == nil || r.DbContext.GetDatabase() == nil {
		return nil, sql.ErrConnDone
	}

	models := make([]AllowedIPModel, 0)

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetAllowedIPsQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			model, scanErr := scanAllowedIP(rows)
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

func (r *Repository) GetByID(id int) (AllowedIPModel, error) {
	var model AllowedIPModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var scanErr error
		model, scanErr = scanAllowedIP(tx.QueryRow(queries.GetAllowedIPByIDQuery, id))
		return scanErr
	})
	if err != nil {
		return AllowedIPModel{}, err
	}

	return model, nil
}

func (r *Repository) Create(tx *sql.Tx, model AllowedIPModel) (AllowedIPModel, error) {
	created, err := scanAllowedIP(tx.QueryRow(
		queries.CreateAllowedIPQuery,
		model.CIDR,
		model.Label,
		model.Enabled,
	))
	if err != nil {
		return AllowedIPModel{}, fmt.Errorf("Create: %w", err)
	}
	return created, nil
}

func (r *Repository) Update(tx *sql.Tx, model AllowedIPModel) (AllowedIPModel, error) {
	updated, err := scanAllowedIP(tx.QueryRow(
		queries.UpdateAllowedIPQuery,
		model.ID,
		model.CIDR,
		model.Label,
		model.Enabled,
	))
	if err != nil {
		return AllowedIPModel{}, fmt.Errorf("Update: %w", err)
	}
	return updated, nil
}

func (r *Repository) Delete(tx *sql.Tx, id int) error {
	result, err := tx.Exec(queries.DeleteAllowedIPQuery, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete rows affected: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
