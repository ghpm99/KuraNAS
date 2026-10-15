package video

import (
	"database/sql"
	"errors"
	"regexp"
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
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertVideoMetadataQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(33, now))
	mock.ExpectCommit()
	err := repo.Db.ExecTx(func(tx *sql.Tx) error {
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestVideoMetadataRepositoryErrorPaths(t *testing.T) {
	repo, mock, db := newVideoMetadataRepoWithMock(t)
	defer db.Close()

	scanErr := errors.New("scan failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertVideoMetadataQuery)).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	err := repo.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertVideoMetadata(tx, VideoMetadataModel{})
		return err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected raw upsert video error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
