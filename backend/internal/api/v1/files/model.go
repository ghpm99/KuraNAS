package files

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"
)

type FileModel struct {
	ID              int
	Name            string
	Path            string
	ParentPath      string
	Type            FileType
	Format          string
	Size            int64
	UpdatedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       sql.NullTime
	LastInteraction sql.NullTime
	LastBackup      sql.NullTime
	CheckSum        string
	Starred         bool
}

type RecentFileModel struct {
	ID         int
	IPAddress  string
	FileID     int
	AccessedAt time.Time
}

func (i *FileDto) ToModel() (FileModel, error) {

	fileModel := FileModel{
		ID:         i.ID,
		Name:       i.Name,
		Path:       i.Path,
		ParentPath: i.ParentPath,
		Type:       i.Type,
		Format:     i.Format,
		Size:       i.Size,
		UpdatedAt:  i.UpdatedAt,
		CreatedAt:  i.CreatedAt,
		CheckSum:   i.CheckSum,
		Starred:    i.Starred,
	}

	deletedAt, err := i.DeletedAt.ParseToNullTime()
	if err != nil {
		return fileModel, err
	}
	fileModel.DeletedAt = deletedAt

	lastInteraction, err := i.LastInteraction.ParseToNullTime()
	if err != nil {
		return fileModel, err
	}
	fileModel.LastInteraction = lastInteraction

	lastBackup, err := i.LastBackup.ParseToNullTime()

	if err != nil {
		return fileModel, err
	}
	fileModel.LastBackup = lastBackup

	return fileModel, nil
}

func (fileModel *FileModel) GetCheckSumFromFile() error {
	file, err := os.Open(fileModel.Path)

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

type SizeReportModel struct {
	Format string
	Total  int
	Size   int64
}

type DuplicateFilesModel struct {
	Name   string
	Size   int64
	Copies int
	Paths  string
}

// Todo: trocar filePath por Path e adicionar coluna de fileId
type ImageMetadataModel struct {
	ID        int
	FileId    int
	Path      string
	Format    string
	Mode      string
	Width     int
	Height    int
	Info      map[string]any
	CreatedAt time.Time
}

type AudioMetadataModel struct {
	ID        int
	FileId    int
	Path      string
	Mime      string
	Info      map[string]any
	Tags      string
	CreatedAt time.Time
}

type VideoMetadataModel struct {
	ID        int
	FileId    int
	Path      string
	Format    string
	Streams   string
	CreatedAt time.Time
}
