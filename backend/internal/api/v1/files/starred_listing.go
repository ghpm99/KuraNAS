package files

import (
	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

type StarredKind string

type StarredSort string

const (
	StarredKindAll     StarredKind = "all"
	StarredKindFolders StarredKind = "folders"
	StarredKindFiles   StarredKind = "files"
	StarredKindMedia   StarredKind = "media"

	StarredSortByName      StarredSort = "name"
	StarredSortByStarredAt StarredSort = "starred_at"
)

type StarredListing struct {
	Kind StarredKind
	Sort StarredSort
}

var DefaultStarredListing = StarredListing{Kind: StarredKindAll, Sort: StarredSortByName}

var starredQueriesBySort = map[StarredSort]string{
	StarredSortByName:      queries.GetStarredFilesQuery,
	StarredSortByStarredAt: queries.GetStarredFilesSortedByStarredAtQuery,
}

var starredKinds = map[StarredKind]bool{
	StarredKindAll:     true,
	StarredKindFolders: true,
	StarredKindFiles:   true,
	StarredKindMedia:   true,
}

// ParseStarredListing converts the raw "kind" and "sort" query values into a
// StarredListing. Empty values select the defaults (all, name); any other
// unknown value makes isValid false.
func ParseStarredListing(rawKind string, rawSort string) (listing StarredListing, isValid bool) {
	listing = DefaultStarredListing
	if rawKind != "" {
		listing.Kind = StarredKind(rawKind)
	}
	if rawSort != "" {
		listing.Sort = StarredSort(rawSort)
	}
	_, isSortKnown := starredQueriesBySort[listing.Sort]
	return listing, starredKinds[listing.Kind] && isSortKnown
}

func (listing StarredListing) query() string {
	return starredQueriesBySort[listing.Sort]
}

func mediaFormats() []string {
	formats := make([]string, 0, len(utils.ImageFormats)+len(utils.VideoFormats)+len(utils.AudioFormats))
	formats = append(formats, utils.ImageFormats...)
	formats = append(formats, utils.VideoFormats...)
	return append(formats, utils.AudioFormats...)
}
