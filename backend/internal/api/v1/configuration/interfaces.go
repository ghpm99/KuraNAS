package configuration

import (
	"database/sql"

	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetSettingsDocument(settingKey string) (string, error)
	UpsertSettingsDocument(tx *sql.Tx, settingKey string, payload string) error
}

type ServiceInterface interface {
	GetSettings() (SettingsDto, error)
	UpdateSettings(request UpdateSettingsRequest) (SettingsDto, error)
	GetTranslationFilePath() (string, error)
	ApplyRuntimeSettings() error
	IsAIImageClassificationEnabled() (bool, error)
}
