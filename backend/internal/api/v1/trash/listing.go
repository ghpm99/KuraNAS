package trash

import (
	"strings"

	queries "nas-go/api/pkg/database/queries/trash"
)

type ListingSortKey string

const (
	SortByDeletedAt ListingSortKey = "deleted_at"
	SortByName      ListingSortKey = "name"
	SortBySize      ListingSortKey = "size"
)

type ListingDirection string

const (
	SortAscending  ListingDirection = "asc"
	SortDescending ListingDirection = "desc"
)

// Listing describes how the trash is listed: its order and an optional
// case-insensitive text that must appear in the item's original path.
type Listing struct {
	SortKey    ListingSortKey
	Direction  ListingDirection
	SearchText string
}

// DefaultListing shows the newest deletions first, unfiltered.
var DefaultListing = Listing{SortKey: SortByDeletedAt, Direction: SortDescending}

var listingQueries = map[ListingSortKey]map[ListingDirection]string{
	SortByDeletedAt: {
		SortAscending:  queries.GetTrashItemsSortedByDeletedAtAscQuery,
		SortDescending: queries.GetTrashItemsSortedByDeletedAtDescQuery,
	},
	SortByName: {
		SortAscending:  queries.GetTrashItemsSortedByNameAscQuery,
		SortDescending: queries.GetTrashItemsSortedByNameDescQuery,
	},
	SortBySize: {
		SortAscending:  queries.GetTrashItemsSortedBySizeAscQuery,
		SortDescending: queries.GetTrashItemsSortedBySizeDescQuery,
	},
}

// ParseListing reads the sort, order and search query parameters. An absent
// sort or order falls back to the default; an unknown one is rejected.
func ParseListing(rawSortKey string, rawDirection string, rawSearchText string) (Listing, bool) {
	listing := DefaultListing
	listing.SearchText = strings.ToLower(strings.TrimSpace(rawSearchText))

	if rawSortKey != "" {
		listing.SortKey = ListingSortKey(rawSortKey)
	}
	if rawDirection != "" {
		listing.Direction = ListingDirection(rawDirection)
	}

	_, isKnown := listingQueries[listing.SortKey][listing.Direction]
	return listing, isKnown
}

func (listing Listing) query() string {
	return listingQueries[listing.SortKey][listing.Direction]
}
