package files

import (
	"nas-go/api/pkg/utils"
	"os"
	"time"
)

type FileType int

const (
	Directory FileType = 1
	File      FileType = 2
)

type FileDto struct {
	ID              int                       `json:"id"`
	Name            string                    `json:"name"`
	Path            string                    `json:"path"`
	Type            FileType                  `json:"type"`
	Format          string                    `json:"format"`
	Size            int64                     `json:"size"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	CreatedAt       time.Time                 `json:"created_at"`
	DeletedAt       utils.Optional[time.Time] `json:"deleted_at"`
	LastInteraction utils.Optional[time.Time] `json:"last_interaction"`
	LastBackup      utils.Optional[time.Time] `json:"last_backup"`
	CheckSum        string                    `json:"check_sum"`
}

func (i *FileModel) ToDto() (FileDto, error) {

	fileDto := FileDto{
		ID:        i.ID,
		Name:      i.Name,
		Path:      i.Path,
		Type:      i.Type,
		Format:    i.Format,
		Size:      i.Size,
		UpdatedAt: i.UpdatedAt,
		CreatedAt: i.CreatedAt,
		CheckSum:  i.CheckSum,
	}

	err := fileDto.DeletedAt.ParseFromNullTime(i.DeletedAt)
	if err != nil {
		return fileDto, err
	}

	err = fileDto.LastInteraction.ParseFromNullTime(i.LastInteraction)
	if err != nil {
		return fileDto, err
	}

	err = fileDto.LastBackup.ParseFromNullTime(i.LastBackup)
	if err != nil {
		return fileDto, err
	}

	return fileDto, nil
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
	Name       string `json:"name"`
	Path       string `json:"path"`
	FileParent int    `json:"file_parent"`
}
