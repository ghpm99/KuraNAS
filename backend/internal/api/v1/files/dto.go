package files

import (
	"os"
	"time"
)

type FileType int

const (
	Directory FileType = 1
	File      FileType = 2
)

type FileDto struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Path            string    `json:"path"`
	Type            FileType  `json:"type"`
	Format          string    `json:"format"`
	Size            int64     `json:"size"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
	DeletedAt       time.Time `json:"deleted_at"`
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

func (fileDto *FileDto) ParseDirEntryToFileDto(entry os.DirEntry) error {
	fileInfo, err := entry.Info()
	if err != nil {
		return err
	}

	if entry.IsDir() {
		fileDto.Type = Directory
	} else {
		fileDto.Type = File
	}

	fileDto.Name = entry.Name()
	fileDto.Size = fileInfo.Size()
	fileDto.UpdatedAt = fileInfo.ModTime()

	return nil
}

type FileFilter struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
