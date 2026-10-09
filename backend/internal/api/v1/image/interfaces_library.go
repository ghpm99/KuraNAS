package image

import (
	"time"

	"nas-go/api/pkg/utils"
)

type LibraryRepositoryInterface interface {
	ListLibraryImages(query LibraryListQuery) ([]LibraryItemModel, error)
	GetLibraryItemCursor(fileID int) (LibraryCursor, error)
	CountLibraryImages(filter LibraryFilter) (int, error)
	ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketModel, error)
	ListLibraryFolders(query LibraryFolderQuery) ([]LibraryFolderModel, error)
	ListLibraryCameraFacets(filter LibraryFilter) ([]LibraryCameraFacetModel, error)
	ListLibraryFormatFacets(filter LibraryFilter) ([]LibraryFormatFacetModel, error)
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

// LibraryFolderRequest asks for one page of direct child folders (with images
// in their subtree) of ParentPath, a client-visible path; empty means the roots.
type LibraryFolderRequest struct {
	ParentPath string
	Page       int
	PageSize   int
}

type LibraryNeighborsRequest struct {
	FileID int
	Filter LibraryFilter
	Count  int
}

type LibraryServiceInterface interface {
	ListLibraryImages(request LibraryListRequest) (LibraryPageDto, error)
	ListLibraryNeighbors(request LibraryNeighborsRequest) (LibraryNeighborsDto, error)
	CountLibraryImages(filter LibraryFilter) (LibraryCountDto, error)
	ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketDto, error)
	ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error)
	ListLibraryCameraFacets(filter LibraryFilter) ([]LibraryCameraFacetDto, error)
	ListLibraryFormatFacets(filter LibraryFilter) ([]LibraryFormatFacetDto, error)
}
