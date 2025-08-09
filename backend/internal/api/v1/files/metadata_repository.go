package files

import (
	"database/sql"
	"fmt"
	"log"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"
	"time"
)

type MetadataRepository struct {
	Db *database.DbContext
}

func NewMetadataRepository(db *database.DbContext) *MetadataRepository {
	return &MetadataRepository{Db: db}
}

func (r *MetadataRepository) GetImageMetadataByID(id int) (ImageMetadataModel, error) {
	var metadata ImageMetadataModel

	// Usa QueryTx para gerenciar o lock de leitura e a transação
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta é movida para dentro desta função anônima
		row := tx.QueryRow(queries.GetImageMetadataByIDQuery, id)

		// Escaneia o resultado
		if err := row.Scan(
			&metadata.ID,
			&metadata.FileId,
			&metadata.Path,
			&metadata.Format,
			&metadata.Mode,
			&metadata.Width,
			&metadata.Height,
			&metadata.DPIX,
			&metadata.DPIY,
			&metadata.XResolution,
			&metadata.YResolution,
			&metadata.ResolutionUnit,
			&metadata.Orientation,
			&metadata.Compression,
			&metadata.Photometric,
			&metadata.ColorSpace,
			&metadata.ComponentsConfig,
			&metadata.ICCProfile,
			&metadata.Make,
			&metadata.Model,
			&metadata.Software,
			&metadata.LensModel,
			&metadata.SerialNumber,
			&metadata.DateTime,
			&metadata.DateTimeOriginal,
			&metadata.DateTimeDigitized,
			&metadata.SubSecTime,
			&metadata.ExposureTime,
			&metadata.FNumber,
			&metadata.ISO,
			&metadata.ShutterSpeed,
			&metadata.ApertureValue,
			&metadata.BrightnessValue,
			&metadata.ExposureBias,
			&metadata.MeteringMode,
			&metadata.Flash,
			&metadata.FocalLength,
			&metadata.WhiteBalance,
			&metadata.ExposureProgram,
			&metadata.MaxApertureValue,
			&metadata.GPSLatitude,
			&metadata.GPSLongitude,
			&metadata.GPSAltitude,
			&metadata.GPSDate,
			&metadata.GPSTime,
			&metadata.ImageDescription,
			&metadata.UserComment,
			&metadata.Copyright,
			&metadata.Artist,
			&metadata.CreatedAt,
		); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return metadata, fmt.Errorf("falha ao obter metadados da imagem: %w", err)
	}

	return metadata, nil
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
		metadata.DPIX,
		metadata.DPIY,
		metadata.XResolution,
		metadata.YResolution,
		metadata.ResolutionUnit,
		metadata.Orientation,
		metadata.Compression,
		metadata.Photometric,
		metadata.ColorSpace,
		metadata.ComponentsConfig,
		metadata.ICCProfile,
		metadata.Make,
		metadata.Model,
		metadata.Software,
		metadata.LensModel,
		metadata.SerialNumber,
		metadata.DateTime,
		metadata.DateTimeOriginal,
		metadata.DateTimeDigitized,
		metadata.SubSecTime,
		metadata.ExposureTime,
		metadata.FNumber,
		metadata.ISO,
		metadata.ShutterSpeed,
		metadata.ApertureValue,
		metadata.BrightnessValue,
		metadata.ExposureBias,
		metadata.MeteringMode,
		metadata.Flash,
		metadata.FocalLength,
		metadata.WhiteBalance,
		metadata.ExposureProgram,
		metadata.MaxApertureValue,
		metadata.GPSLatitude,
		metadata.GPSLongitude,
		metadata.GPSAltitude,
		metadata.GPSDate,
		metadata.GPSTime,
		metadata.ImageDescription,
		metadata.UserComment,
		metadata.Copyright,
		metadata.Artist,
		time.Now(),
	}

	var row *sql.Row
	row = tx.QueryRow(queries.UpsertImageMetadataQuery, args...)

	err := row.Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

func (r *MetadataRepository) DeleteImageMetadata(id int) error {
	// Usa ExecTx para gerenciar o lock de escrita e a transação
	// A lógica de execução da query é movida para dentro desta função anônima
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.DeleteImageMetadataQuery, id)
		if err != nil {
			return fmt.Errorf("falha ao executar query de exclusão: %w", err)
		}
		return nil
	})

	if err != nil {
		// Se houver um erro, ele será propagado por ExecTx
		return fmt.Errorf("falha ao deletar metadados da imagem: %w", err)
	}

	return nil
}

func (r *MetadataRepository) GetAudioMetadataByID(id int) (AudioMetadataModel, error) {
	var metadata AudioMetadataModel

	err := r.Db.QueryTx(func(tx *sql.Tx) error {

		row := tx.QueryRow(queries.GetAudioMetadataByIDQuery, id)

		if err := row.Scan(
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
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return metadata, fmt.Errorf("falha ao obter metadados de audio: %w", err)
	}

	return metadata, nil
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
	row = tx.QueryRow(queries.UpsertAudioMetadataQuery, args...)

	err := row.Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}
func (r *MetadataRepository) DeleteAudioMetadata(id int) error {
	// Usa ExecTx para gerenciar o lock de escrita e a transação
	// A lógica de execução da query é movida para dentro desta função anônima
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.DeleteAudioMetadataQuery, id)
		if err != nil {
			return fmt.Errorf("falha ao executar query de exclusão: %w", err)
		}
		return nil
	})

	if err != nil {
		// Se houver um erro, ele será propagado por ExecTx
		return fmt.Errorf("falha ao deletar metadados de audio: %w", err)
	}

	return nil
}

func (r *MetadataRepository) GetVideoMetadataByID(id int) (VideoMetadataModel, error) {
	var metadata VideoMetadataModel

	// Usa QueryTx para gerenciar o lock de leitura e a transação
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		// A lógica de consulta é movida para dentro desta função anônima
		row := tx.QueryRow(queries.GetVideoMetadataByIDQuery, id)

		// Escaneia o resultado
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

	row = tx.QueryRow(queries.UpsertVideoMetadataQuery, args...)

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
	// Usa ExecTx para gerenciar o lock de escrita e a transação
	// A lógica de execução da query é movida para dentro desta função anônima
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.DeleteVideoMetadataQuery, id)
		if err != nil {
			return fmt.Errorf("falha ao executar query de exclusão: %w", err)
		}
		return nil
	})

	if err != nil {
		// Se houver um erro, ele será propagado por ExecTx
		return fmt.Errorf("falha ao deletar metadados do vídeo: %w", err)
	}

	return nil
}
