package files

import "time"

type FileDto struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Path            string    `json:"path"`
	Format          string    `json:"format"`
	Size            int64     `json:"size"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
	LastInteraction time.Time `json:"last_interaction"`
	LastBackup      time.Time `json:"last_backup"`
}

func (i *FileModel) ToDto() FileDto {
	return FileDto{
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
