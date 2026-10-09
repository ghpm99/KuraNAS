package files

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const searchDateLayout = "2006-01-02"

var searchKindsByName = map[string]FileSearchKind{
	string(SearchKindFolder):   SearchKindFolder,
	string(SearchKindDocument): SearchKindDocument,
	string(SearchKindImage):    SearchKindImage,
	string(SearchKindAudio):    SearchKindAudio,
	string(SearchKindVideo):    SearchKindVideo,
	string(SearchKindArchive):  SearchKindArchive,
	string(SearchKindOther):    SearchKindOther,
}

func parseSearchKinds(rawKinds []string) ([]FileSearchKind, bool) {
	kinds := make([]FileSearchKind, 0, len(rawKinds))
	for _, rawKind := range rawKinds {
		kind, isKnown := searchKindsByName[rawKind]
		if !isKnown {
			return nil, false
		}
		kinds = append(kinds, kind)
	}
	return kinds, true
}

func parseSearchDate(rawDate string) (*time.Time, bool) {
	if rawDate == "" {
		return nil, true
	}
	parsedDate, err := time.ParseInLocation(searchDateLayout, rawDate, time.UTC)
	if err != nil {
		return nil, false
	}
	return &parsedDate, true
}

func parseSearchSize(rawSize string) (*int64, bool) {
	if rawSize == "" {
		return nil, true
	}
	size, err := strconv.ParseInt(rawSize, 10, 64)
	if err != nil || size < 0 {
		return nil, false
	}
	return &size, true
}

func parseSearchTier(rawTier string) (string, bool) {
	switch rawTier {
	case "", TierHot, TierCold:
		return rawTier, true
	}
	return "", false
}

func parseSearchOnlyStarred(rawStarred string) (bool, bool) {
	switch rawStarred {
	case "", "false":
		return false, true
	case "true":
		return true, true
	}
	return false, false
}

func parseSearchSort(rawSort string) (FileSearchSort, bool) {
	switch sort := FileSearchSort(rawSort); sort {
	case SearchSortRelevance, SearchSortName, SearchSortSize, SearchSortModified:
		return sort, true
	case "":
		return SearchSortRelevance, true
	}
	return "", false
}

func parseSearchOrder(rawOrder string) (FileSearchOrder, bool) {
	switch order := FileSearchOrder(rawOrder); order {
	case SearchOrderAscending, SearchOrderDescending, "":
		return order, true
	}
	return "", false
}

func isSearchRangeInverted[Bound int64 | time.Time](lower *Bound, upper *Bound, isAfter func(Bound, Bound) bool) bool {
	return lower != nil && upper != nil && isAfter(*lower, *upper)
}

func parseSearchFilter(c *gin.Context) (FileSearchFilter, string) {
	kinds, isKindsValid := parseSearchKinds(c.QueryArray("kind"))
	if !isKindsValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_KIND"
	}
	modifiedFrom, isFromValid := parseSearchDate(c.Query("modified_from"))
	modifiedTo, isToValid := parseSearchDate(c.Query("modified_to"))
	if !isFromValid || !isToValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_DATE"
	}
	minSize, isMinValid := parseSearchSize(c.Query("min_size"))
	maxSize, isMaxValid := parseSearchSize(c.Query("max_size"))
	if !isMinValid || !isMaxValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_SIZE"
	}
	isDateRangeInverted := isSearchRangeInverted(modifiedFrom, modifiedTo, time.Time.After)
	isSizeRangeInverted := isSearchRangeInverted(minSize, maxSize, func(lower int64, upper int64) bool { return lower > upper })
	if isDateRangeInverted || isSizeRangeInverted {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_RANGE"
	}
	tier, isTierValid := parseSearchTier(c.Query("tier"))
	if !isTierValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_TIER"
	}
	onlyStarred, isStarredValid := parseSearchOnlyStarred(c.Query("starred"))
	if !isStarredValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_STARRED"
	}
	sort, isSortValid := parseSearchSort(c.Query("sort"))
	if !isSortValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_SORT"
	}
	order, isOrderValid := parseSearchOrder(c.Query("order"))
	if !isOrderValid {
		return FileSearchFilter{}, "ERROR_SEARCH_INVALID_ORDER"
	}
	return FileSearchFilter{
		Kinds:        kinds,
		ModifiedFrom: modifiedFrom,
		ModifiedTo:   modifiedTo,
		MinSize:      minSize,
		MaxSize:      maxSize,
		Tier:         tier,
		OnlyStarred:  onlyStarred,
		Sort:         sort,
		Order:        order,
	}, ""
}
