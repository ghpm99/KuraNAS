package image

import "time"

type LibraryRepositoryInterface interface {
	ListLibraryImages(query LibraryListQuery) ([]LibraryItemModel, error)
	CountLibraryImages(filter LibraryFilter) (int, error)
	ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketModel, error)
}

// LibraryListRequest is a gallery page request as parsed from the HTTP layer.
// Page is only used by the offset orderings; keyset paging uses Cursor.
type LibraryListRequest struct {
	Filter      LibraryFilter
	Sort        LibrarySort
	Order       LibrarySortOrder
	Cursor      *LibraryCursor
	TakenBefore *time.Time
	Page        int
	PageSize    int
}

type LibraryServiceInterface interface {
	ListLibraryImages(request LibraryListRequest) (LibraryPageDto, error)
	CountLibraryImages(filter LibraryFilter) (LibraryCountDto, error)
	ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketDto, error)
}
