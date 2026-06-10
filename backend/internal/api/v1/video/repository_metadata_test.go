package video

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/video"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newVideoMetadataRepoWithMock(t *testing.T) (*VideoMetadataRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewVideoMetadataRepository(database.NewDbContext(db)), mock, db
}

func TestVideoMetadataRepositorySuccessPaths(t *testing.T) {
	repo, mock, db := newVideoMetadataRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	if repo == nil || repo.GetDbContext() == nil {
		t.Fatalf("expected initialized video metadata repository")
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

func TestVideoMetadataRepositoryErrorPaths(t *testing.T) {
	repo, mock, db := newVideoMetadataRepoWithMock(t)
	defer db.Close()

	scanErr := errors.New("scan failed")
	execErr := errors.New("exec failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoMetadataByIDQuery)).
		WithArgs(1).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	_, err := repo.GetVideoMetadataByID(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter metadados do vídeo") {
		t.Fatalf("expected wrapped video error, got: %v", err)
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
