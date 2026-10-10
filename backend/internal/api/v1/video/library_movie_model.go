package video

type LibraryMovieSort string

const (
	LibraryMovieSortName   LibraryMovieSort = "name"
	LibraryMovieSortRecent LibraryMovieSort = "recent"
)

func ParseLibraryMovieSort(rawSort string) (LibraryMovieSort, bool) {
	switch LibraryMovieSort(rawSort) {
	case "", LibraryMovieSortName:
		return LibraryMovieSortName, true
	case LibraryMovieSortRecent:
		return LibraryMovieSortRecent, true
	}
	return "", false
}

type LibraryMoviesRequest struct {
	Sort     LibraryMovieSort
	Page     int
	PageSize int
}
