package files

import (
	"errors"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestRepositoryCountStarredFilesByKindReturnsCounts(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CountStarredFilesByKindQuery)).
		WithArgs(pq.Array(mediaFormats())).
		WillReturnRows(sqlmock.NewRows([]string{"total", "folders", "files", "media"}).AddRow(6, 1, 2, 3))
	mock.ExpectRollback()

	counts, err := repo.CountStarredFilesByKind()
	if err != nil || counts != (StarredCountsDto{Total: 6, Folders: 1, Files: 2, Media: 3}) {
		t.Fatalf("unexpected result: %+v err=%v", counts, err)
	}
}

func TestRepositoryCountStarredFilesByKindPropagatesQueryError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	queryFailure := errors.New("query failed")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CountStarredFilesByKindQuery)).
		WithArgs(pq.Array(mediaFormats())).
		WillReturnError(queryFailure)
	mock.ExpectRollback()

	if _, err := repo.CountStarredFilesByKind(); !errors.Is(err, queryFailure) {
		t.Fatalf("expected query failure, got %v", err)
	}
}
