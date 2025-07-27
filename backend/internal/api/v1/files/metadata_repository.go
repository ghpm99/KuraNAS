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

func (r *MetadataRepository) GetImageMetadataByID(id int) (ImageMetadataModel, error) {
	var metadata ImageMetadataModel
	var infoStr string

	err := r.Db.QueryRow(queries.GetImageMetadataByIDQuery, id).Scan(
		&metadata.ID,
		&metadata.FileId,
		&metadata.Path,
		&metadata.Format,
		&metadata.Mode,
		&metadata.Width,
		&metadata.Height,
		&infoStr,
		&metadata.CreatedAt,
	)

	if err != nil {
		return metadata, err
	}

	if err = json.Unmarshal([]byte(infoStr), &metadata.Info); err != nil {
		return metadata, err
	}

	return metadata, nil
}

func (r *MetadataRepository) UpsertImageMetadata(tx *sql.Tx, metadata ImageMetadataModel) (ImageMetadataModel, error) {
	var id int
	var createdAt time.Time

	infoJson, err := json.Marshal(metadata.Info)
	if err != nil {
		return metadata, err
	}

	queryArgs := []any{
		metadata.ID,
		metadata.FileId,
		metadata.Path,
		metadata.Format,
		metadata.Mode,
		metadata.Width,
		metadata.Height,
		infoJson,
		time.Now(),
	}
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(
			queries.UpsertImageMetadataQuery,
			queryArgs,
		)
	} else {
		row = r.Db.QueryRow(
			queries.UpsertImageMetadataQuery,
			queryArgs,
		)
	}

	err = row.Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

func (r *MetadataRepository) DeleteImageMetadata(id int) error {
	_, err := r.Db.Exec(queries.DeleteImageMetadataQuery, id)
	return err
}

func (r *MetadataRepository) GetAudioMetadataByID(id int) (AudioMetadataModel, error) {
	var metadata AudioMetadataModel

	var infoStr string

	err := r.Db.QueryRow(queries.GetAudioMetadataByIDQuery, id).Scan(
		&metadata.ID,
		&metadata.FileId,
		&metadata.Path,
		&metadata.Mime,
		&infoStr,
		&metadata.Tags,
		&metadata.CreatedAt,
	)

	if err != nil {
		return metadata, err
	}

	if err = json.Unmarshal([]byte(infoStr), &metadata.Info); err != nil {
		return metadata, err
	}

	return metadata, err
}

func (r *MetadataRepository) UpsertAudioMetadata(tx *sql.Tx, metadata AudioMetadataModel) (AudioMetadataModel, error) {
	var id int
	var createdAt time.Time

	infoJson, err := json.Marshal(metadata.Info)
	if err != nil {
		return metadata, err
	}

	args := []any{
		metadata.FileId,
		metadata.Path,
		metadata.Mime,
		infoJson,
		metadata.Tags,
		time.Now(),
	}

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(queries.UpsertAudioMetadataQuery, args...)
	} else {
		row = r.Db.QueryRow(queries.UpsertAudioMetadataQuery, args...)
	}

	err = row.Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

func (r *MetadataRepository) DeleteAudioMetadata(id int) error {
	_, err := r.Db.Exec(queries.DeleteAudioMetadataQuery, id)
	return err
}

func (r *MetadataRepository) GetVideoMetadataByID(id int) (VideoMetadataModel, error) {
	var metadata VideoMetadataModel

	err := r.Db.QueryRow(queries.GetVideoMetadataByIDQuery, id).Scan(
		&metadata.ID,
		&metadata.FileId,
		&metadata.Path,
		&metadata.Format,
		&metadata.Streams,
		&metadata.CreatedAt,
	)

	return metadata, err
}

func (r *MetadataRepository) UpsertVideoMetadata(tx *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error) {
	var id int
	var createdAt time.Time

	args := []any{
		metadata.FileId,
		metadata.Path,
		metadata.Format,
		metadata.Streams,
		time.Now(),
	}

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(queries.UpsertVideoMetadataQuery, args...)
	} else {
		row = r.Db.QueryRow(queries.UpsertVideoMetadataQuery, args...)
	}

	err := row.Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

func (r *MetadataRepository) DeleteVideoMetadata(id int) error {
	_, err := r.Db.Exec(queries.DeleteVideoMetadataQuery, id)
	return err
}
