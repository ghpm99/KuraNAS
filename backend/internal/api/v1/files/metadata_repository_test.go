package files

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMetadataRepoWithMock(t *testing.T) (*MetadataRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewMetadataRepository(database.NewDbContext(db)), mock, db
}

func TestMetadataRepositorySuccessPaths(t *testing.T) {
	repo, mock, db := newMetadataRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	if repo == nil || repo.Db == nil {
		t.Fatalf("expected initialized metadata repository")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageMetadataByIDQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "file_id", "path", "format", "mode", "width", "height", "dpi_x", "dpi_y", "x_resolution",
			"y_resolution", "resolution_unit", "orientation", "compression", "photometric", "color_space",
			"components_configuration", "icc_profile", "make", "model", "software", "lens_model", "serial_number",
			"datetime", "datetime_original", "datetime_digitized", "subsec_time", "exposure_time", "f_number", "iso",
			"shutter_speed", "aperture_value", "brightness_value", "exposure_bias", "metering_mode", "flash", "focal_length",
			"white_balance", "exposure_program", "max_aperture_value", "gps_latitude", "gps_longitude", "gps_altitude",
			"gps_date", "gps_time", "image_description", "user_comment", "copyright", "artist", "created_at",
		}).AddRow(
			1, 10, "/i.jpg", "jpeg", "rgb", 100, 80, 72.0, 72.0, 72.0, 72.0, 2.0, 1.0, 1.0, 2.0, 1.0,
			"cfg", "icc", "mk", "md", "sw", "lens", "sn", "2026", "2026", "2026", "1", 0.1, 2.8, 200.0,
			3.0, 2.0, 1.0, 0.0, 5.0, 0.0, 35.0, 0.0, 1.0, 2.0, -23.5, -46.6, 700.0, "2026-01-01",
			"12:00:00", "desc", "comment", "cpr", "art", now,
		))
	mock.ExpectRollback()
	imageMeta, err := repo.GetImageMetadataByID(1)
	if err != nil || imageMeta.ID != 1 {
		t.Fatalf("GetImageMetadataByID failed meta=%+v err=%v", imageMeta, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAudioMetadataByIDQuery)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "file_id", "path", "mime", "length", "bitrate", "sample_rate", "channels", "bitrate_mode", "encoder_info",
			"bit_depth", "title", "artist", "album", "album_artist", "track_number", "genre", "composer", "year",
			"recording_date", "encoder", "publisher", "original_release_date", "original_artist", "lyricist", "lyrics", "created_at",
		}).AddRow(
			2, 20, "/a.mp3", "audio/mpeg", 120.0, 320, 44100, 2, 1, "enc", 16, "title", "artist", "album",
			"album artist", "1", "genre", "composer", "2026", "2026-01-01", "encoder", "publisher", "2025-01-01",
			"original", "lyricist", "lyrics", now,
		))
	mock.ExpectRollback()
	audioMeta, err := repo.GetAudioMetadataByID(2)
	if err != nil || audioMeta.ID != 2 {
		t.Fatalf("GetAudioMetadataByID failed meta=%+v err=%v", audioMeta, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoMetadataByIDQuery)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{
			"file_id", "path", "format_name", "size", "duration", "width", "height", "frame_rate", "nb_frames", "bit_rate",
			"codec_name", "codec_long_name", "pix_fmt", "level", "profile", "aspect_ratio", "audio_codec", "audio_channels",
			"audio_sample_rate", "audio_bit_rate", "created_at",
		}).AddRow(
			30, "/v.mp4", "mov,mp4", "1000", "60", 1920, 1080, 30.0, 1800, "2500", "h264",
			"H.264 / AVC", "yuv420p", 4, "main", "16:9", "aac", 2, "44100", "128000", now,
		))
	mock.ExpectRollback()
	videoMeta, err := repo.GetVideoMetadataByID(3)
	if err != nil || videoMeta.FileId != 30 {
		t.Fatalf("GetVideoMetadataByID failed meta=%+v err=%v", videoMeta, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertImageMetadataQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(11, now))
	mock.ExpectCommit()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		upserted, err := repo.UpsertImageMetadata(tx, ImageMetadataModel{FileId: 10, Path: "/i.jpg"})
		if err != nil {
			return err
		}
		if upserted.ID != 11 {
			t.Fatalf("expected image metadata ID 11")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpsertImageMetadata failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertAudioMetadataQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(22, now))
	mock.ExpectCommit()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		upserted, err := repo.UpsertAudioMetadata(tx, AudioMetadataModel{FileId: 20, Path: "/a.mp3"})
		if err != nil {
			return err
		}
		if upserted.ID != 22 {
			t.Fatalf("expected audio metadata ID 22")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpsertAudioMetadata failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertVideoMetadataQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(33, now))
	mock.ExpectCommit()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		upserted, err := repo.UpsertVideoMetadata(tx, VideoMetadataModel{FileId: 30, Path: "/v.mp4"})
		if err != nil {
			return err
		}
		if upserted.ID != 33 {
			t.Fatalf("expected video metadata ID 33")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpsertVideoMetadata failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteImageMetadataQuery)).
		WithArgs(11).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.DeleteImageMetadata(11); err != nil {
		t.Fatalf("DeleteImageMetadata failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAudioMetadataQuery)).
		WithArgs(22).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.DeleteAudioMetadata(22); err != nil {
		t.Fatalf("DeleteAudioMetadata failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteVideoMetadataQuery)).
		WithArgs(33).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.DeleteVideoMetadata(33); err != nil {
		t.Fatalf("DeleteVideoMetadata failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestMetadataRepositoryErrorPaths(t *testing.T) {
	repo, mock, db := newMetadataRepoWithMock(t)
	defer db.Close()

	scanErr := errors.New("scan failed")
	execErr := errors.New("exec failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageMetadataByIDQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	_, err := repo.GetImageMetadataByID(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter metadados da imagem") {
		t.Fatalf("expected wrapped image error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAudioMetadataByIDQuery)).
		WithArgs(1).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	_, err = repo.GetAudioMetadataByID(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter metadados de audio") {
		t.Fatalf("expected wrapped audio error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoMetadataByIDQuery)).
		WithArgs(1).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	_, err = repo.GetVideoMetadataByID(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter metadados do vídeo") {
		t.Fatalf("expected wrapped video error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertImageMetadataQuery)).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertImageMetadata(tx, ImageMetadataModel{})
		return err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected raw upsert image error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertAudioMetadataQuery)).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertAudioMetadata(tx, AudioMetadataModel{})
		return err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected raw upsert audio error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertVideoMetadataQuery)).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertVideoMetadata(tx, VideoMetadataModel{})
		return err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected raw upsert video error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteImageMetadataQuery)).
		WithArgs(1).
		WillReturnError(execErr)
	mock.ExpectRollback()
	err = repo.DeleteImageMetadata(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao deletar metadados da imagem") {
		t.Fatalf("expected wrapped delete image error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAudioMetadataQuery)).
		WithArgs(1).
		WillReturnError(execErr)
	mock.ExpectRollback()
	err = repo.DeleteAudioMetadata(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao deletar metadados de audio") {
		t.Fatalf("expected wrapped delete audio error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteVideoMetadataQuery)).
		WithArgs(1).
		WillReturnError(execErr)
	mock.ExpectRollback()
	err = repo.DeleteVideoMetadata(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao deletar metadados do vídeo") {
		t.Fatalf("expected wrapped delete video error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
