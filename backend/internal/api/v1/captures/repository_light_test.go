package captures

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/captures"

	"github.com/DATA-DOG/go-sqlmock"
)

func newCapturesRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestRepositoryGetDbContext(t *testing.T) {
	repo, _, db := newCapturesRepoWithMock(t)
	defer db.Close()

	if repo.GetDbContext() == nil {
		t.Fatal("expected non-nil db context")
	}
}

func TestRepositoryCreateCapture(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertCaptureQuery)).
		WithArgs("test", "video.ts", "/data/capturas/test/video.ts", "hls", "video/mp2t", int64(1024), now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		result, err := repo.CreateCapture(tx, CaptureModel{
			Name:      "test",
			FileName:  "video.ts",
			FilePath:  "/data/capturas/test/video.ts",
			MediaType: "hls",
			MimeType:  "video/mp2t",
			Size:      1024,
			CreatedAt: now,
		})
		if err != nil {
			return err
		}
		if result.ID != 1 {
			t.Fatalf("expected ID 1, got %d", result.ID)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("CreateCapture failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryCreateCaptureError(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertCaptureQuery)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.CreateCapture(tx, CaptureModel{Name: "fail"})
		return err
	})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryGetCaptures(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "file_name", "file_path", "media_type", "mime_type", "size", "created_at"}).
		AddRow(1, "test", "video.ts", "/path", "hls", "video/mp2t", 1024, now).
		AddRow(2, "other", "stream.mp4", "/path2", "dash", "video/mp4", 2048, now)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetCapturesQuery)).WillReturnRows(rows)
	mock.ExpectRollback()

	result, err := repo.GetCaptures(CaptureFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetCapturesError(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetCapturesQuery)).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	_, err := repo.GetCaptures(CaptureFilter{}, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryGetCaptureByID(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetCaptureByIDQuery)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "file_name", "file_path", "media_type", "mime_type", "size", "created_at"}).
			AddRow(5, "test", "video.ts", "/path", "hls", "video/mp2t", 1024, now))
	mock.ExpectRollback()

	result, err := repo.GetCaptureByID(5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != 5 {
		t.Fatalf("expected ID 5, got %d", result.ID)
	}
	if result.Name != "test" {
		t.Fatalf("expected name test, got %s", result.Name)
	}
}

func TestRepositoryGetCaptureByIDNotFound(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetCaptureByIDQuery)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.GetCaptureByID(99)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryDeleteCapture(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteCaptureQuery)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeleteCapture(tx, 1)
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRepositoryDeleteCaptureNotFound(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteCaptureQuery)).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeleteCapture(tx, 99)
	})

	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestRepositoryDeleteCaptureExecError(t *testing.T) {
	repo, mock, db := newCapturesRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteCaptureQuery)).
		WithArgs(1).
		WillReturnError(errors.New("exec failed"))
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeleteCapture(tx, 1)
	})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewServiceReturnsNonNil(t *testing.T) {
	repo, _, db := newCapturesRepoWithMock(t)
	defer db.Close()

	service := NewService(repo, nil, nil)
	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestNewRepositoryReturnsNonNil(t *testing.T) {
	dbCtx := database.NewDbContext(nil)
	repo := NewRepository(dbCtx)
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}
