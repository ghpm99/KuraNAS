package files

import "time"

type FileModel struct {
	ID              int
	Name            string
	Path            string
	Format          string
	Size            int64
	UpdatedAt       time.Time
	CreatedAt       time.Time
	LastInteraction time.Time
	LastBackup      time.Time
}

func (i *FileDto) ToModel() FileModel {
	return FileModel{
		ID:              i.ID,
		Name:            i.Name,
		Path:            i.Path,
		Format:          i.Format,
		Size:            i.Size,
		UpdatedAt:       i.UpdatedAt,
		CreatedAt:       i.CreatedAt,
		LastInteraction: i.LastInteraction,
		LastBackup:      i.LastBackup,
	}
}
