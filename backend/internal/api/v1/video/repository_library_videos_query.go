package video

import (
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
)

var libraryVideosOrderFragments = map[LibraryVideoOrdering]string{
	{Sort: LibraryVideoSortRecent, Direction: SortDirectionAscending}:    queries.GetLibraryVideosOrderRecentAscQuery,
	{Sort: LibraryVideoSortRecent, Direction: SortDirectionDescending}:   queries.GetLibraryVideosOrderRecentDescQuery,
	{Sort: LibraryVideoSortName, Direction: SortDirectionAscending}:      queries.GetLibraryVideosOrderNameAscQuery,
	{Sort: LibraryVideoSortName, Direction: SortDirectionDescending}:     queries.GetLibraryVideosOrderNameDescQuery,
	{Sort: LibraryVideoSortSize, Direction: SortDirectionAscending}:      queries.GetLibraryVideosOrderSizeAscQuery,
	{Sort: LibraryVideoSortSize, Direction: SortDirectionDescending}:     queries.GetLibraryVideosOrderSizeDescQuery,
	{Sort: LibraryVideoSortDuration, Direction: SortDirectionAscending}:  queries.GetLibraryVideosOrderDurationAscQuery,
	{Sort: LibraryVideoSortDuration, Direction: SortDirectionDescending}: queries.GetLibraryVideosOrderDurationDescQuery,
}

func libraryVideosQueryFor(ordering LibraryVideoOrdering) (string, error) {
	orderFragment, isWhitelisted := libraryVideosOrderFragments[ordering]
	if !isWhitelisted {
		return "", fmt.Errorf("unsupported library video ordering: %s %s", ordering.Sort, ordering.Direction)
	}
	return queries.GetLibraryVideosQuery + orderFragment + queries.GetLibraryVideosPageQuery, nil
}
