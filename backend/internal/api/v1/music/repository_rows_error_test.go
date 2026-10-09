package music

import (
	"errors"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/music"

	"github.com/DATA-DOG/go-sqlmock"
)

func failingRows(iterationFailure error) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id"}).AddRow(1).RowError(0, iterationFailure)
}

func TestMusicListingsReturnRowIterationError(t *testing.T) {
	iterationFailure := errors.New("connection reset while streaming rows")

	cases := map[string]struct {
		query string
		call  func(repository *Repository) error
	}{
		"GetMusic": {queries.GetMusicQuery, func(repository *Repository) error {
			_, err := repository.GetMusic(1, 10)
			return err
		}},
		"GetLibraryTracks": {queries.GetLibraryTracksQuery, func(repository *Repository) error {
			_, err := repository.GetLibraryTracks(1, 10)
			return err
		}},
		"GetLibraryFilesByIDs": {queries.GetLibraryFilesByIDsQuery, func(repository *Repository) error {
			_, err := repository.GetLibraryFilesByIDs([]int{1})
			return err
		}},
		"GetPlaylistTracks": {queries.GetPlaylistTracksQuery, func(repository *Repository) error {
			_, err := repository.GetPlaylistTracks(1, 1, 10)
			return err
		}},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			repository, mock, db := newMusicRepoWithMock(t)
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(testCase.query)).WillReturnRows(failingRows(iterationFailure))
			mock.ExpectRollback()

			if err := testCase.call(repository); !errors.Is(err, iterationFailure) {
				t.Fatalf("expected iteration error, got %v", err)
			}
		})
	}
}
