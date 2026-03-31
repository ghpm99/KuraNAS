package libraries

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/libraries"

	"github.com/DATA-DOG/go-sqlmock"
)

func newLibrariesRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestLibrariesRepositoryUpsert_InsertOrUpdate(t *testing.T) {
	repo, mock, db := newLibrariesRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertLibraryQuery)).
		WithArgs("images", "/data/Imagens").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category", "path", "created_at", "updated_at"}).
			AddRow(1, "images", "/data/Imagens", now, now))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, upsertErr := repo.Upsert(tx, LibraryModel{Category: LibraryCategoryImages, Path: "/data/Imagens"})
		return upsertErr
	})
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLibrariesRepositoryGetAll(t *testing.T) {
	repo, mock, db := newLibrariesRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLibrariesQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category", "path", "created_at", "updated_at"}).
			AddRow(1, "documents", "/data/Documentos", now, now).
			AddRow(2, "images", "/data/Imagens", now, now))
	mock.ExpectRollback()

	models, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLibrariesRepositoryGetByCategory(t *testing.T) {
	repo, mock, db := newLibrariesRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLibraryByCategoryQuery)).
		WithArgs("music").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category", "path", "created_at", "updated_at"}).
			AddRow(7, "music", "/data/Musicas", now, now))
	mock.ExpectRollback()

	model, err := repo.GetByCategory(LibraryCategoryMusic)
	if err != nil {
		t.Fatalf("GetByCategory returned error: %v", err)
	}
	if model.Category != LibraryCategoryMusic {
		t.Fatalf("expected category music, got %s", model.Category)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
