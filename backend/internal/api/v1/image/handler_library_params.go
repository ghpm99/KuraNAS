package image

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxLibraryNameQueryLength   = 200
	defaultLibraryNeighborCount = 20
	maxLibraryNeighborCount     = 50
)

var (
	errInvalidLibraryCategory = errors.New("invalid library category")
	errInvalidLibraryFormat   = errors.New("invalid library format")
	errInvalidLibraryDate     = errors.New("invalid library date")
	errInvalidLibrarySort     = errors.New("invalid library sort")
	errInvalidLibraryOrder    = errors.New("invalid library order")
	errInvalidLibraryStarred  = errors.New("invalid library starred")
	errInvalidLibraryQuery    = errors.New("invalid library query")
	errInvalidLibraryFolder   = errors.New("invalid library folder")
	errInvalidLibraryFileID   = errors.New("invalid library file id")
	errInvalidLibraryCount    = errors.New("invalid library neighbors count")
	errLibraryKeysetOnly      = errors.New("library cursor requires taken_at desc ordering")
)

var libraryErrorMessageKeys = map[error]string{
	ErrInvalidLibraryCursor:   "ERROR_IMAGE_LIBRARY_INVALID_CURSOR",
	errInvalidLibraryCategory: "ERROR_IMAGE_LIBRARY_INVALID_CATEGORY",
	errInvalidLibraryFormat:   "ERROR_IMAGE_LIBRARY_INVALID_FORMAT",
	errInvalidLibraryDate:     "ERROR_IMAGE_LIBRARY_INVALID_DATE",
	errInvalidLibrarySort:     "ERROR_IMAGE_LIBRARY_INVALID_SORT",
	errInvalidLibraryOrder:    "ERROR_IMAGE_LIBRARY_INVALID_ORDER",
	errInvalidLibraryStarred:  "ERROR_IMAGE_LIBRARY_INVALID_STARRED",
	errInvalidLibraryQuery:    "ERROR_IMAGE_LIBRARY_INVALID_QUERY",
	errInvalidLibraryFolder:   "ERROR_IMAGE_LIBRARY_INVALID_FOLDER",
	errInvalidLibraryCount:    "ERROR_IMAGE_LIBRARY_INVALID_COUNT",
	errLibraryKeysetOnly:      "ERROR_IMAGE_LIBRARY_KEYSET_REQUIRES_DATE_SORT",
}

func libraryErrorMessageKey(err error) string {
	for knownError, messageKey := range libraryErrorMessageKeys {
		if errors.Is(err, knownError) {
			return messageKey
		}
	}
	return "ERROR_INVALID_REQUEST"
}

func parseLibraryFilter(c *gin.Context) (LibraryFilter, error) {
	filter := LibraryFilter{
		NameQuery: strings.ToLower(strings.TrimSpace(c.Query("q"))),
		Camera:    strings.TrimSpace(c.Query("camera")),
		Folder:    strings.TrimSpace(c.Query("folder")),
	}
	if len(filter.NameQuery) > maxLibraryNameQueryLength {
		return LibraryFilter{}, errInvalidLibraryQuery
	}

	var err error
	if filter.Categories, err = parseLibraryCategories(c.QueryArray("category")); err != nil {
		return LibraryFilter{}, err
	}
	if filter.Formats, err = parseLibraryFormats(c.QueryArray("format")); err != nil {
		return LibraryFilter{}, err
	}
	if filter.OnlyStarred, err = parseLibraryStarred(c.Query("starred")); err != nil {
		return LibraryFilter{}, err
	}
	if filter.TakenFrom, err = parseLibraryDate(c.Query("taken_from"), false); err != nil {
		return LibraryFilter{}, err
	}
	if filter.TakenTo, err = parseLibraryDate(c.Query("taken_to"), true); err != nil {
		return LibraryFilter{}, err
	}
	return filter, nil
}

func parseLibraryListRequest(c *gin.Context) (LibraryListRequest, error) {
	filter, err := parseLibraryFilter(c)
	if err != nil {
		return LibraryListRequest{}, err
	}

	sort, err := parseLibrarySort(c.Query("sort"))
	if err != nil {
		return LibraryListRequest{}, err
	}
	order, err := parseLibrarySortOrder(c.Query("order"), sort)
	if err != nil {
		return LibraryListRequest{}, err
	}

	request := LibraryListRequest{Filter: filter, Sort: sort, Order: order}
	rawCursor := c.Query("cursor")
	rawTakenBefore := c.Query("taken_before")
	if (rawCursor != "" || rawTakenBefore != "") && !IsKeysetOrdering(sort, order) {
		return LibraryListRequest{}, errLibraryKeysetOnly
	}
	if rawCursor != "" {
		cursor, err := DecodeLibraryCursor(rawCursor)
		if err != nil {
			return LibraryListRequest{}, err
		}
		request.Cursor = &cursor
	}
	if request.TakenBefore, err = parseLibraryDate(rawTakenBefore, false); err != nil {
		return LibraryListRequest{}, err
	}
	return request, nil
}

func parseLibraryCategories(rawCategories []string) ([]ClassificationCategory, error) {
	categories := make([]ClassificationCategory, 0, len(rawCategories))
	for _, rawCategory := range rawCategories {
		category, isKnown := ParseClassificationCategory(rawCategory)
		if !isKnown {
			return nil, errInvalidLibraryCategory
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func parseLibraryFormats(rawFormats []string) ([]string, error) {
	formats := make([]string, 0, len(rawFormats))
	for _, rawFormat := range rawFormats {
		format := strings.ToLower(strings.TrimSpace(rawFormat))
		if !strings.HasPrefix(format, ".") {
			format = "." + format
		}
		if !slices.Contains(utils.ImageFormats, format) {
			return nil, errInvalidLibraryFormat
		}
		formats = append(formats, format)
	}
	return formats, nil
}

func parseLibraryStarred(rawStarred string) (bool, error) {
	if rawStarred == "" {
		return false, nil
	}
	isStarred, err := strconv.ParseBool(rawStarred)
	if err != nil {
		return false, errInvalidLibraryStarred
	}
	return isStarred, nil
}

func parseLibrarySort(rawSort string) (LibrarySort, error) {
	switch LibrarySort(rawSort) {
	case "":
		return LibrarySortTakenAt, nil
	case LibrarySortTakenAt, LibrarySortName, LibrarySortSize:
		return LibrarySort(rawSort), nil
	default:
		return "", errInvalidLibrarySort
	}
}

func parseLibrarySortOrder(rawOrder string, sort LibrarySort) (LibrarySortOrder, error) {
	switch LibrarySortOrder(rawOrder) {
	case "":
		return defaultLibrarySortOrder(sort), nil
	case LibrarySortOrderAsc, LibrarySortOrderDesc:
		return LibrarySortOrder(rawOrder), nil
	default:
		return "", errInvalidLibraryOrder
	}
}

func defaultLibrarySortOrder(sort LibrarySort) LibrarySortOrder {
	if sort == LibrarySortName {
		return LibrarySortOrderAsc
	}
	return LibrarySortOrderDesc
}

const libraryDateOnlyLayout = "2006-01-02"

func parseLibraryDate(rawDate string, isEndOfDayWhenDateOnly bool) (*time.Time, error) {
	if rawDate == "" {
		return nil, nil
	}
	if dateOnly, err := time.ParseInLocation(libraryDateOnlyLayout, rawDate, time.UTC); err == nil {
		if isEndOfDayWhenDateOnly {
			dateOnly = dateOnly.AddDate(0, 0, 1).Add(-time.Microsecond)
		}
		return &dateOnly, nil
	}
	instant, err := time.Parse(time.RFC3339, rawDate)
	if err != nil {
		return nil, errInvalidLibraryDate
	}
	instant = instant.UTC()
	return &instant, nil
}

func parseLibraryNeighborCount(rawCount string) (int, error) {
	if rawCount == "" {
		return defaultLibraryNeighborCount, nil
	}
	count, err := strconv.Atoi(rawCount)
	if err != nil || count < 1 || count > maxLibraryNeighborCount {
		return 0, errInvalidLibraryCount
	}
	return count, nil
}
