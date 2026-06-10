package video

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/video"
)

// VideoMetadataRepository is the write-side for the video_metadata complement table.
type VideoMetadataRepository struct {
	Db *database.DbContext
}

func NewVideoMetadataRepository(db *database.DbContext) *VideoMetadataRepository {
	return &VideoMetadataRepository{Db: db}
}

func (r *VideoMetadataRepository) GetDbContext() *database.DbContext {
	return r.Db
}

func (r *VideoMetadataRepository) GetVideoMetadataByID(id int) (VideoMetadataModel, error) {
	var metadata VideoMetadataModel

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetVideoMetadataByIDQuery, id)

		if err := row.Scan(
			&metadata.FileId,
			&metadata.Path,
			&metadata.FormatName,
			&metadata.Size,
			&metadata.Duration,
			&metadata.Width,
			&metadata.Height,
			&metadata.FrameRate,
			&metadata.NbFrames,
			&metadata.BitRate,
			&metadata.CodecName,
			&metadata.CodecLongName,
			&metadata.PixFmt,
			&metadata.Level,
			&metadata.Profile,
			&metadata.AspectRatio,
			&metadata.AudioCodec,
			&metadata.AudioChannels,
			&metadata.AudioSampleRate,
			&metadata.AudioBitRate,
			&metadata.CreatedAt,
		); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return metadata, fmt.Errorf("falha ao obter metadados do vídeo: %w", err)
	}

	return metadata, nil
}

func (r *VideoMetadataRepository) UpsertVideoMetadata(tx *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error) {
	var id int
	var createdAt time.Time

	args := []any{
		metadata.FileId,
		metadata.Path,
		metadata.FormatName,
		metadata.Size,
		metadata.Duration,
		metadata.Width,
		metadata.Height,
		metadata.FrameRate,
		metadata.NbFrames,
		metadata.BitRate,
		metadata.CodecName,
		metadata.CodecLongName,
		metadata.PixFmt,
		metadata.Level,
		metadata.Profile,
		metadata.AspectRatio,
		metadata.AudioCodec,
		metadata.AudioChannels,
		metadata.AudioSampleRate,
		metadata.AudioBitRate,
		time.Now(),
	}

	row := tx.QueryRow(queries.UpsertVideoMetadataQuery, args...)

	err := row.Scan(&id, &createdAt)
	if err != nil {
		log.Println("Error scanning video metadata:", err)
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

func (r *VideoMetadataRepository) DeleteVideoMetadata(id int) error {
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.DeleteVideoMetadataQuery, id)
		if err != nil {
			return fmt.Errorf("falha ao executar query de exclusão: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("falha ao deletar metadados do vídeo: %w", err)
	}

	return nil
}
