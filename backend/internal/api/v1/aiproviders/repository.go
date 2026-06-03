package aiproviders

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/aiproviders"
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

func scanProvider(scan func(dest ...any) error) (ProviderModel, error) {
	var m ProviderModel
	var name string
	var paramsRaw []byte

	if err := scan(&m.ID, &name, &m.Enabled, &m.Model, &m.BaseURL, &m.Priority, &paramsRaw, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return ProviderModel{}, err
	}

	params, err := decodeParams(paramsRaw)
	if err != nil {
		return ProviderModel{}, fmt.Errorf("decode params: %w", err)
	}

	m.Name = ProviderName(name)
	m.Params = params
	return m, nil
}

func (r *Repository) GetAll() ([]ProviderModel, error) {
	var providers []ProviderModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetAIProvidersQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			m, scanErr := scanProvider(rows.Scan)
			if scanErr != nil {
				return scanErr
			}
			providers = append(providers, m)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}

	return providers, nil
}

func (r *Repository) GetByName(name ProviderName) (ProviderModel, error) {
	var m ProviderModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var scanErr error
		m, scanErr = scanProvider(tx.QueryRow(queries.GetAIProviderByNameQuery, string(name)).Scan)
		return scanErr
	})

	if err != nil {
		return ProviderModel{}, fmt.Errorf("GetByName: %w", err)
	}

	return m, nil
}

func (r *Repository) InsertIfAbsent(model ProviderModel) error {
	paramsJSON, err := encodeParams(model.Params)
	if err != nil {
		return fmt.Errorf("InsertIfAbsent encode params: %w", err)
	}

	err = r.DbContext.QueryTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(
			queries.InsertAIProviderIfAbsentQuery,
			string(model.Name),
			model.Enabled,
			model.Model,
			model.BaseURL,
			model.Priority,
			paramsJSON,
		)
		return execErr
	})

	if err != nil {
		return fmt.Errorf("InsertIfAbsent: %w", err)
	}

	return nil
}

func (r *Repository) Update(model ProviderModel) (ProviderModel, error) {
	paramsJSON, err := encodeParams(model.Params)
	if err != nil {
		return ProviderModel{}, fmt.Errorf("Update encode params: %w", err)
	}

	var updated ProviderModel
	err = r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var scanErr error
		updated, scanErr = scanProvider(tx.QueryRow(
			queries.UpdateAIProviderQuery,
			string(model.Name),
			model.Enabled,
			model.Model,
			model.BaseURL,
			model.Priority,
			paramsJSON,
		).Scan)
		return scanErr
	})

	if err != nil {
		return ProviderModel{}, fmt.Errorf("Update: %w", err)
	}

	return updated, nil
}
