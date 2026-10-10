package video

type LibraryVideoSort string

type SortDirection string

const (
	LibraryVideoSortRecent   LibraryVideoSort = "recent"
	LibraryVideoSortName     LibraryVideoSort = "name"
	LibraryVideoSortSize     LibraryVideoSort = "size"
	LibraryVideoSortDuration LibraryVideoSort = "duration"
)

const (
	SortDirectionAscending  SortDirection = "asc"
	SortDirectionDescending SortDirection = "desc"
)

type LibraryVideoOrdering struct {
	Sort      LibraryVideoSort
	Direction SortDirection
}

type LibraryVideosRequest struct {
	Ordering    LibraryVideoOrdering
	Page        int
	PageSize    int
	SearchQuery string
}

func defaultDirectionFor(sort LibraryVideoSort) SortDirection {
	if sort == LibraryVideoSortName {
		return SortDirectionAscending
	}
	return SortDirectionDescending
}

func ParseLibraryVideoOrdering(rawSort string, rawDirection string) (LibraryVideoOrdering, bool) {
	sort := LibraryVideoSort(rawSort)
	switch sort {
	case "":
		sort = LibraryVideoSortRecent
	case LibraryVideoSortRecent, LibraryVideoSortName, LibraryVideoSortSize, LibraryVideoSortDuration:
	default:
		return LibraryVideoOrdering{}, false
	}

	direction := SortDirection(rawDirection)
	switch direction {
	case "":
		direction = defaultDirectionFor(sort)
	case SortDirectionAscending, SortDirectionDescending:
	default:
		return LibraryVideoOrdering{}, false
	}
	return LibraryVideoOrdering{Sort: sort, Direction: direction}, true
}
