package files

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryToggleFileStarredReturnsNewValue(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ToggleFileStarredQuery)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"starred"}).AddRow(true))
	mock.ExpectCommit()

	isStarred, err := repo.ToggleFileStarred(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isStarred {
		t.Fatalf("expected starred true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryToggleFileStarredWithoutMatchingRowReturnsErrNoRows(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ToggleFileStarredQuery)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"starred"}))
	mock.ExpectRollback()

	_, err := repo.ToggleFileStarred(7)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRepositoryToggleFileStarredPropagatesQueryError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	queryFailure := errors.New("query failed")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ToggleFileStarredQuery)).
		WithArgs(7).
		WillReturnError(queryFailure)
	mock.ExpectRollback()

	_, err := repo.ToggleFileStarred(7)
	if !errors.Is(err, queryFailure) {
		t.Fatalf("expected query failure, got %v", err)
	}
}
