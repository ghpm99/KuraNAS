package storageroots

import (
	"time"

	"nas-go/api/internal/roots"
)

type StorageRootModel struct {
	ID        int
	Path      string
	Label     string
	Enabled   bool
	CreatedAt time.Time
}

func (model *StorageRootModel) ToDto() StorageRootDto {
	return StorageRootDto{
		ID:        model.ID,
		Path:      model.Path,
		Label:     model.Label,
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt,
	}
}

func (model *StorageRootModel) toRegistryRoot() roots.Root {
	return roots.Root{
		ID:      model.ID,
		Path:    model.Path,
		Label:   model.Label,
		Enabled: model.Enabled,
	}
}
