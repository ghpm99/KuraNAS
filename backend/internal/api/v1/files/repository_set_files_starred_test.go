package files

import (
	"errors"
	"reflect"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestRepositorySetFilesStarredReturnsUpdatedIDs(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFilesStarredQuery)).
		WithArgs(pq.Array([]int{3, 5, 9}), true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3).AddRow(9))
	mock.ExpectCommit()

	updatedIDs, err := repo.SetFilesStarred([]int{3, 5, 9}, true)
	if err != nil || !reflect.DeepEqual(updatedIDs, []int{3, 9}) {
		t.Fatalf("unexpected result: ids=%v err=%v", updatedIDs, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositorySetFilesStarredWithoutMatchesReturnsEmptySlice(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFilesStarredQuery)).
		WithArgs(pq.Array([]int{1}), false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()

	updatedIDs, err := repo.SetFilesStarred([]int{1}, false)
	if err != nil || updatedIDs == nil || len(updatedIDs) != 0 {
		t.Fatalf("unexpected result: ids=%#v err=%v", updatedIDs, err)
	}
}

func TestRepositorySetFilesStarredPropagatesQueryError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	queryFailure := errors.New("query failed")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFilesStarredQuery)).
		WithArgs(pq.Array([]int{1}), true).
		WillReturnError(queryFailure)
	mock.ExpectRollback()

	if _, err := repo.SetFilesStarred([]int{1}, true); !errors.Is(err, queryFailure) {
		t.Fatalf("expected query failure, got %v", err)
	}
}

func TestRepositorySetFilesStarredPropagatesScanError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SetFilesStarredQuery)).
		WithArgs(pq.Array([]int{1}), true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("not-a-number"))
	mock.ExpectRollback()

	if _, err := repo.SetFilesStarred([]int{1}, true); err == nil {
		t.Fatal("expected scan error")
	}
}
