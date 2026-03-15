package configuration

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/configuration"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetSettingsDocument(settingKey string) (string, error)
	UpsertSettingsDocument(tx *sql.Tx, settingKey string, payload string) error
}

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(dbContext *database.DbContext) *Repository {
	return &Repository{DbContext: dbContext}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) GetSettingsDocument(settingKey string) (string, error) {
	var key string
	var payload string

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetSettingQuery, settingKey).Scan(&key, &payload)
	})
	if err != nil {
		return "", fmt.Errorf("falha ao buscar documento de configuracao: %w", err)
	}

	return payload, nil
}

func (r *Repository) UpsertSettingsDocument(tx *sql.Tx, settingKey string, payload string) error {
	if _, err := tx.Exec(queries.UpsertSettingQuery, settingKey, payload); err != nil {
		return fmt.Errorf("falha ao persistir documento de configuracao: %w", err)
	}
	return nil
}
