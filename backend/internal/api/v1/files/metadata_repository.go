package files

import (
	"database/sql"
	"encoding/json"
	queries "nas-go/api/pkg/database/queries/file"
	"time"
)

type MetadataRepository struct {
	Db *sql.DB
}

func NewMetadataRepository(db *sql.DB) *MetadataRepository {
	return &MetadataRepository{Db: db}
}

func (r *MetadataRepository) CreateImageMetadata(metadata ImageMetadataModel) (ImageMetadataModel, error) {
	var id int
	var createdAt time.Time
	infoJson, _ := json.Marshal(metadata.Info)
	err := r.Db.QueryRow(
		queries.InsertImageMetadataQuery,
		metadata.FilePath,
		metadata.Format,
		metadata.Mode,
		metadata.Width,
		metadata.Height,
		infoJson,
		time.Now(),
	).Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}
	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

func (r *MetadataRepository) GetImageMetadataByID(id int) (ImageMetadataModel, error) {
	var metadata ImageMetadataModel
	var infoStr string

	err := r.Db.QueryRow(queries.GetImageMetadataByIDQuery, id).Scan(
		&metadata.ID,
		&metadata.FilePath,
		&metadata.Format,
		&metadata.Mode,
		&metadata.Width,
		&metadata.Height,
		&infoStr,
		&metadata.CreatedAt,
	)

	if err = json.Unmarshal([]byte(infoStr), &metadata.Info); err != nil {
		return metadata, err
	}

	return metadata, nil
}

func (r *MetadataRepository) UpdateImageMetadata(metadata ImageMetadataModel) (ImageMetadataModel, error) {
	_, err := r.Db.Exec(
		queries.UpdateImageMetadataQuery,
		metadata.FilePath,
		metadata.Format,
		metadata.Mode,
		metadata.Width,
		metadata.Height,
		metadata.Info,
		metadata.ID,
	)
	return metadata, err
}

func (r *MetadataRepository) DeleteImageMetadata(id int) error {
	_, err := r.Db.Exec(queries.DeleteImageMetadataQuery, id)
	return err
}
