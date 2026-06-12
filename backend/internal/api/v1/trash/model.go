package trash

import "time"

// TrashItemModel is the DB shape of one item living in the trash directory.
type TrashItemModel struct {
	ID           int
	OriginalPath string
	TrashPath    string
	Size         int64
	DeletedAt    time.Time
}

func (model *TrashItemModel) ToDto() TrashItemDto {
	return TrashItemDto{
		ID:           model.ID,
		OriginalPath: model.OriginalPath,
		Size:         model.Size,
		DeletedAt:    model.DeletedAt,
	}
}
