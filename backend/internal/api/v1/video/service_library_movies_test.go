package video

import (
	"errors"
	"testing"
)

func TestListLibraryMoviesPagesWithOneExtraRowAndMapsDtos(t *testing.T) {
	var receivedSort LibraryMovieSort
	var receivedLimit, receivedOffset int
	repository := &videoRepoMock{listLibraryMoviesFn: func(sort LibraryMovieSort, limit int, offset int) ([]VideoFileModel, error) {
		receivedSort, receivedLimit, receivedOffset = sort, limit, offset
		return []VideoFileModel{{ID: 1, Name: "a.mkv"}, {ID: 2, Name: "b.mkv"}, {ID: 3, Name: "c.mkv"}}, nil
	}}
	service := &Service{Repository: repository}

	moviePage, err := service.ListLibraryMovies(LibraryMoviesRequest{Sort: LibraryMovieSortRecent, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedSort != LibraryMovieSortRecent || receivedLimit != 3 || receivedOffset != 2 {
		t.Fatalf("unexpected window sort=%s limit=%d offset=%d", receivedSort, receivedLimit, receivedOffset)
	}
	if len(moviePage.Items) != 2 || !moviePage.Pagination.HasNext || moviePage.Items[0].ID != 1 {
		t.Fatalf("unexpected page %+v", moviePage)
	}
}

func TestListLibraryMoviesPropagatesRepositoryError(t *testing.T) {
	repository := &videoRepoMock{listLibraryMoviesFn: func(LibraryMovieSort, int, int) ([]VideoFileModel, error) {
		return nil, errors.New("boom")
	}}
	service := &Service{Repository: repository}

	if _, err := service.ListLibraryMovies(LibraryMoviesRequest{Sort: LibraryMovieSortName, Page: 1, PageSize: 5}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseLibraryMovieSortAcceptsOnlyKnownValues(t *testing.T) {
	for rawSort, expected := range map[string]LibraryMovieSort{"": LibraryMovieSortName, "name": LibraryMovieSortName, "recent": LibraryMovieSortRecent} {
		sort, isValid := ParseLibraryMovieSort(rawSort)
		if !isValid || sort != expected {
			t.Fatalf("sort %q: got %q valid=%v", rawSort, sort, isValid)
		}
	}
	if _, isValid := ParseLibraryMovieSort("size"); isValid {
		t.Fatal("expected unknown sort to be invalid")
	}
}
