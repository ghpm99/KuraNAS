package files

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

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
