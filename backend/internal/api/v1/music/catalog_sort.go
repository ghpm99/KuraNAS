package music

type CatalogSortField string

const (
	CatalogSortByTracks CatalogSortField = "tracks"
	CatalogSortByName   CatalogSortField = "name"
	CatalogSortByRecent CatalogSortField = "recent"
	CatalogSortByYear   CatalogSortField = "year"
)

const (
	catalogOrderAscending  = "asc"
	catalogOrderDescending = "desc"
)

type CatalogSort struct {
	Field        CatalogSortField
	IsDescending bool
}

func DefaultCatalogSort() CatalogSort {
	return CatalogSort{Field: CatalogSortByTracks, IsDescending: true}
}

func (sort CatalogSort) normalized() CatalogSort {
	if sort.Field == "" {
		return DefaultCatalogSort()
	}
	return sort
}

func ParseCatalogSort(rawField string, rawOrder string, isYearAllowed bool) (CatalogSort, bool) {
	field := CatalogSortField(rawField)
	switch field {
	case "":
		field = CatalogSortByTracks
	case CatalogSortByTracks, CatalogSortByName, CatalogSortByRecent:
	case CatalogSortByYear:
		if !isYearAllowed {
			return CatalogSort{}, false
		}
	default:
		return CatalogSort{}, false
	}

	switch rawOrder {
	case catalogOrderAscending:
		return CatalogSort{Field: field, IsDescending: false}, true
	case catalogOrderDescending:
		return CatalogSort{Field: field, IsDescending: true}, true
	case "":
		return CatalogSort{Field: field, IsDescending: field != CatalogSortByName}, true
	default:
		return CatalogSort{}, false
	}
}
