package image

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetImagesReturnsRowIterationError(t *testing.T) {
	repo, mock, db := newImageRepoWithMock(t)
	defer db.Close()

	iterationFailure := errors.New("connection reset while streaming rows")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(getImagesQueryByGroup(ImageGroupByDate))).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).RowError(0, iterationFailure))
	mock.ExpectRollback()

	_, err := repo.GetImages(1, 10, ImageGroupByDate)
	if !errors.Is(err, iterationFailure) {
		t.Fatalf("expected iteration error, got %v", err)
	}
}
