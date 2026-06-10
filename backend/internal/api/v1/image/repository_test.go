package image

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/image"

	"github.com/DATA-DOG/go-sqlmock"
)

func newImageRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestImageRepositoryGetDbContext(t *testing.T) {
	db, _, sqlDB := newImageRepoWithMock(t)
	defer sqlDB.Close()
	if db.GetDbContext() == nil {
		t.Fatal("expected non-nil DbContext")
	}
}

func TestGetImageMetadataByID_Success(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()
	now := time.Now()

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
			"gps_date", "gps_time", "image_description", "user_comment", "copyright", "artist",
			"classification_category", "classification_confidence", "classification_suggested_name", "created_at",
		}).AddRow(
			1, 10, "/i.jpg", "jpeg", "rgb", 100, 80, 72.0, 72.0, 72.0, 72.0, 2.0, 1.0, 1.0, 2.0, 1.0,
			"cfg", "icc", "mk", "md", "sw", "lens", "sn", "2026", "2026", "2026", "1", 0.1, 2.8, 200.0,
			3.0, 2.0, 1.0, 0.0, 5.0, 0.0, 35.0, 0.0, 1.0, 2.0, -23.5, -46.6, 700.0, "2026-01-01",
			"12:00:00", "desc", "comment", "cpr", "art", "photo", 0.91, "rikka", now,
		))
	mock.ExpectRollback()

	meta, err := repo.GetImageMetadataByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.ID != 1 {
		t.Fatalf("expected ID 1, got %d", meta.ID)
	}
	if meta.Classification.Category != ClassificationCategoryPhoto {
		t.Fatalf("expected photo classification, got %s", meta.Classification.Category)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetImageMetadataByID_Error(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageMetadataByIDQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // too few columns → scan error
	mock.ExpectRollback()

	_, err := repo.GetImageMetadataByID(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter metadados da imagem") {
		t.Fatalf("expected wrapped error, got: %v", err)
	}
}

func TestUpsertImageMetadata_Success(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertImageMetadataQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(11, now))
	mock.ExpectCommit()

	err := repo.Db.ExecTx(func(tx *sql.Tx) error {
		upserted, err := repo.UpsertImageMetadata(tx, MetadataModel{FileId: 10, Path: "/i.jpg"})
		if err != nil {
			return err
		}
		if upserted.ID != 11 {
			t.Fatalf("expected ID 11, got %d", upserted.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpsertImageMetadata failed: %v", err)
	}
}

func TestUpsertImageMetadata_Error(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()
	scanErr := errors.New("scan failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertImageMetadataQuery)).
		WillReturnError(scanErr)
	mock.ExpectRollback()

	err := repo.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertImageMetadata(tx, MetadataModel{})
		return err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected raw upsert error, got: %v", err)
	}
}

func TestDeleteImageMetadata_Success(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteImageMetadataQuery)).
		WithArgs(11).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.DeleteImageMetadata(11); err != nil {
		t.Fatalf("DeleteImageMetadata failed: %v", err)
	}
}

func TestDeleteImageMetadata_Error(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()
	execErr := errors.New("exec failed")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteImageMetadataQuery)).
		WithArgs(1).
		WillReturnError(execErr)
	mock.ExpectRollback()

	err := repo.DeleteImageMetadata(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao deletar metadados da imagem") {
		t.Fatalf("expected wrapped delete error, got: %v", err)
	}
}
