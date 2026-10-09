package image

import "time"

type LibrarySort string

const (
	LibrarySortTakenAt LibrarySort = "taken_at"
	LibrarySortName    LibrarySort = "name"
	LibrarySortSize    LibrarySort = "size"
)

type LibrarySortOrder string

const (
	LibrarySortOrderDesc LibrarySortOrder = "desc"
	LibrarySortOrderAsc  LibrarySortOrder = "asc"
)

// LibraryFilter holds the combinable gallery filters. Zero values mean the
// filter is not applied. Folder lists direct children only; it is a client-visible
// path in requests and an absolute disk path once the service resolved it.
// AlbumID restricts the listing to the photos of one user album; it is never
// read from the gallery query string, only set by the album service.
type LibraryFilter struct {
	NameQuery               string
	ContentQuery            string
	MustMatchNameAndContent bool
	Categories              []ClassificationCategory
	OnlyStarred             bool
	Formats                 []string
	TakenFrom               *time.Time
	TakenTo                 *time.Time
	Camera                  string
	Folder                  string
	AlbumID                 int
}

// LibraryListQuery is one page request against the gallery listing.
// Cursor, NewerThan and TakenBefore are only honored for the taken_at/desc ordering;
// NewerThan lists the items listed before that position, nearest first.
type LibraryListQuery struct {
	Filter      LibraryFilter
	Sort        LibrarySort
	Order       LibrarySortOrder
	Cursor      *LibraryCursor
	NewerThan   *LibraryCursor
	TakenBefore *time.Time
	Limit       int
	Offset      int
}

type LibraryItemModel struct {
	FileID     int
	Name       string
	Path       string
	ParentPath string
	Format     string
	Size       int64
	Width      int
	Height     int
	TakenAt    *time.Time
	Category   string
	Starred    bool
	IsCold     bool
	UpdatedAt  time.Time
}

type LibraryCameraFacetModel struct {
	Camera string
	Count  int
}

type LibraryFormatFacetModel struct {
	Format string
	Count  int
}

type LibraryTimelineBucketModel struct {
	Year  int
	Month int
	Count int
}

func IsKeysetOrdering(sort LibrarySort, order LibrarySortOrder) bool {
	return sort == LibrarySortTakenAt && order == LibrarySortOrderDesc
}
