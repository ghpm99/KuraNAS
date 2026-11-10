package trash

import (
	"path/filepath"
	"time"
)

// TrashItemModel is the DB shape of one item living in the trash directory.
type TrashItemModel struct {
	ID           int
	OriginalPath string
	TrashPath    string
	Size         int64
	DeletedAt    time.Time
	IsDirectory  bool
}

// ToDto converts the item to its transport shape; retentionDays defines when
// the purger will remove it for good.
func (model *TrashItemModel) ToDto(retentionDays int) TrashItemDto {
	return TrashItemDto{
		ID:           model.ID,
		Name:         filepath.Base(model.OriginalPath),
		OriginalPath: model.OriginalPath,
		Size:         model.Size,
		DeletedAt:    model.DeletedAt,
		ExpiresAt:    model.DeletedAt.AddDate(0, 0, retentionDays),
		IsDirectory:  model.IsDirectory,
	}
}
