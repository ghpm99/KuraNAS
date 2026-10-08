package dav

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/dav"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestColdFileRepositoryListsColdFilesOfParent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()
	repository := NewColdFileRepository(database.NewDbContext(db))
	modTime := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListColdFilesByParentPathQuery)).
		WithArgs("/data").
		WillReturnRows(sqlmock.NewRows([]string{"name", "path", "physical_path", "size", "updated_at"}).
			AddRow("a.pdf", "/data/a.pdf", "/cold/a.pdf", 10, modTime))
	mock.ExpectRollback()

	coldFiles, err := repository.ListColdFilesByParentPath("/data")

	if err != nil || len(coldFiles) != 1 || coldFiles[0].PhysicalPath != "/cold/a.pdf" || coldFiles[0].Size != 10 {
		t.Fatalf("unexpected result: %+v err=%v", coldFiles, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestColdFileRepositoryWrapsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()
	repository := NewColdFileRepository(database.NewDbContext(db))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListColdFilesByParentPathQuery)).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repository.ListColdFilesByParentPath("/data"); err == nil {
		t.Fatalf("expected error")
	}
}
