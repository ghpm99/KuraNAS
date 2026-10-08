package video

import (
	"errors"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/video"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetVideosReturnsRowIterationError(t *testing.T) {
	repo, mock, db := newVideoRepoWithMock(t)
	defer db.Close()

	iterationFailure := errors.New("connection reset while streaming rows")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideosQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).RowError(0, iterationFailure))
	mock.ExpectRollback()

	_, err := repo.GetVideos(1, 10)
	if !errors.Is(err, iterationFailure) {
		t.Fatalf("expected iteration error, got %v", err)
	}
}
