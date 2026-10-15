package files

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositorySetFileStarredReturnsStoredValue(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFileStarredQuery)).
		WithArgs(7, false).
		WillReturnRows(sqlmock.NewRows([]string{"starred"}).AddRow(false))
	mock.ExpectCommit()

	isStarred, err := repo.SetFileStarred(7, false)
	if err != nil || isStarred {
		t.Fatalf("unexpected result: value=%v err=%v", isStarred, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositorySetFileStarredWithoutMatchingRowReturnsErrNoRows(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFileStarredQuery)).
		WithArgs(7, true).
		WillReturnRows(sqlmock.NewRows([]string{"starred"}))
	mock.ExpectRollback()

	_, err := repo.SetFileStarred(7, true)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRepositorySetFileStarredPropagatesQueryError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	queryFailure := errors.New("query failed")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFileStarredQuery)).
		WithArgs(7, true).
		WillReturnError(queryFailure)
	mock.ExpectRollback()

	_, err := repo.SetFileStarred(7, true)
	if !errors.Is(err, queryFailure) {
		t.Fatalf("expected query failure, got %v", err)
	}
}
