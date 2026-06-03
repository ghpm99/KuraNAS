package aiproviders

import (
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetAll() ([]ProviderModel, error)
	GetByName(name ProviderName) (ProviderModel, error)
	InsertIfAbsent(model ProviderModel) error
	Update(model ProviderModel) (ProviderModel, error)
}

type ServiceInterface interface {
	GetProviders() ([]ProviderDto, error)
	UpdateProvider(name ProviderName, dto UpdateProviderDto) (ProviderDto, error)
	EnsureDefaults() error
	GetProviderModels() ([]ProviderModel, error)
	SetOnChange(fn func())
}
