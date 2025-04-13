package files

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"nas-go/api/pkg/utils"
	"os"
	"time"
)

type FileModel struct {
	ID              int
	Name            string
	Path            string
	Type            FileType
	Format          string
	Size            int64
	UpdatedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       utils.Optional[time.Time]
	LastInteraction utils.Optional[time.Time]
	LastBackup      utils.Optional[time.Time]
	CheckSum        string
}

func (i *FileDto) ToModel() FileModel {
	return FileModel{
		ID:              i.ID,
		Name:            i.Name,
		Path:            i.Path,
		Type:            i.Type,
		Format:          i.Format,
		Size:            i.Size,
		UpdatedAt:       i.UpdatedAt,
		CreatedAt:       i.CreatedAt,
		DeletedAt:       i.DeletedAt,
		LastInteraction: i.LastInteraction,
		LastBackup:      i.LastBackup,
		CheckSum:        i.CheckSum,
	}
}

func (fileModel *FileModel) getCheckSumFromFile() error {
	file, err := os.Open(fileModel.Path)

	if fileModel.Size > (1 * 1024 * 1024 * 1024) {
		return errors.New("arquivo muito grande")
	}
	if err != nil {
		return err
	}

	defer file.Close()

	h := sha256.New()

	if _, err := io.Copy(h, file); err != nil {
		return err
	}

	checkSumBytes := h.Sum(nil)
	checkSumString := fmt.Sprintf("%x", checkSumBytes)

	fmt.Printf("Check sum %s, tamanho %d\n", checkSumString, len(checkSumString))

	return nil
}
