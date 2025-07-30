package files

import (
	"database/sql"
	"log"
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
	var m ImageMetadataModel
	err := r.Db.QueryRow(queries.GetImageMetadataByIDQuery, id).Scan(
		&m.ID,
		&m.FileId,
		&m.Path,
		&m.Format,
		&m.Mode,
		&m.Width,
		&m.Height,
		&m.CaptureDate,
		&m.Software,
		&m.Make,
		&m.Model,
		&m.LensModel,
		&m.ISO,
		&m.ExposureTime,
		&m.DPIX,
		&m.DPIY,
		&m.ICCProfile,
		&m.GPSLatitude,
		&m.GPSLongitude,
		&m.CreatedAt,
	)
	return m, err
}

func (r *MetadataRepository) UpsertImageMetadata(tx *sql.Tx, metadata ImageMetadataModel) (ImageMetadataModel, error) {
	var id int
	var createdAt time.Time

	args := []any{
		metadata.FileId,
		metadata.Path,
		metadata.Format,
		metadata.Mode,
		metadata.Width,
		metadata.Height,
		metadata.CaptureDate,
		metadata.Software,
		metadata.Make,
		metadata.Model,
		metadata.LensModel,
		metadata.ISO,
		metadata.ExposureTime,
		metadata.DPIX,
		metadata.DPIY,
		metadata.ICCProfile,
		metadata.GPSLatitude,
		metadata.GPSLongitude,
		time.Now(),
	}

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(queries.UpsertImageMetadataQuery, args...)
	} else {
		row = r.Db.QueryRow(queries.UpsertImageMetadataQuery, args...)
	}

	err := row.Scan(&id, &createdAt)
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

	err := r.Db.QueryRow(queries.GetAudioMetadataByIDQuery, id).Scan(
		&metadata.ID,
		&metadata.FileId,
		&metadata.Path,
		&metadata.Mime,
		&metadata.Length,
		&metadata.Bitrate,
		&metadata.SampleRate,
		&metadata.Channels,
		&metadata.BitrateMode,
		&metadata.EncoderInfo,
		&metadata.BitDepth,
		&metadata.Title,
		&metadata.Artist,
		&metadata.Album,
		&metadata.AlbumArtist,
		&metadata.TrackNumber,
		&metadata.Genre,
		&metadata.Composer,
		&metadata.Year,
		&metadata.RecordingDate,
		&metadata.Encoder,
		&metadata.Publisher,
		&metadata.OriginalReleaseDate,
		&metadata.OriginalArtist,
		&metadata.Lyricist,
		&metadata.Lyrics,
		&metadata.CreatedAt,
	)
	return metadata, err
}

func (r *MetadataRepository) UpsertAudioMetadata(tx *sql.Tx, metadata AudioMetadataModel) (AudioMetadataModel, error) {
	var id int
	var createdAt time.Time

	args := []any{
		metadata.FileId,
		metadata.Path,
		metadata.Mime,
		metadata.Length,
		metadata.Bitrate,
		metadata.SampleRate,
		metadata.Channels,
		metadata.BitrateMode,
		metadata.EncoderInfo,
		metadata.BitDepth,
		metadata.Title,
		metadata.Artist,
		metadata.Album,
		metadata.AlbumArtist,
		metadata.TrackNumber,
		metadata.Genre,
		metadata.Composer,
		metadata.Year,
		metadata.RecordingDate,
		metadata.Encoder,
		metadata.Publisher,
		metadata.OriginalReleaseDate,
		metadata.OriginalArtist,
		metadata.Lyricist,
		metadata.Lyrics,
		time.Now(),
	}

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(queries.UpsertAudioMetadataQuery, args...)
	} else {
		row = r.Db.QueryRow(queries.UpsertAudioMetadataQuery, args...)
	}

	err := row.Scan(&id, &createdAt)
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
	)
	if err != nil {
		return metadata, err
	}

	return metadata, err
}

func (r *MetadataRepository) UpsertVideoMetadata(tx *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error) {
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

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(queries.UpsertVideoMetadataQuery, args...)
	} else {
		row = r.Db.QueryRow(queries.UpsertVideoMetadataQuery, args...)
	}

	err := row.Scan(&id, &createdAt)
	if err != nil {
		log.Println("Error scanning video metadata:", err)
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
