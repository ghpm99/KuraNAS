package files

import "time"

const MinSearchQueryLength = 2

type FileSearchKind string

const (
	SearchKindFolder   FileSearchKind = "folder"
	SearchKindDocument FileSearchKind = "document"
	SearchKindImage    FileSearchKind = "image"
	SearchKindAudio    FileSearchKind = "audio"
	SearchKindVideo    FileSearchKind = "video"
	SearchKindArchive  FileSearchKind = "archive"
	SearchKindOther    FileSearchKind = "other"
)

type FileSearchSort string

const (
	SearchSortRelevance FileSearchSort = "relevance"
	SearchSortName      FileSearchSort = "name"
	SearchSortSize      FileSearchSort = "size"
	SearchSortModified  FileSearchSort = "modified"
)

type FileSearchOrder string

const (
	SearchOrderAscending  FileSearchOrder = "asc"
	SearchOrderDescending FileSearchOrder = "desc"
)

type FileSearchScope int

const (
	SearchScopeGlobal FileSearchScope = iota
	SearchScopeDescendants
	SearchScopeChildren
)

type FileSearchFilter struct {
	Kinds        []FileSearchKind
	ModifiedFrom *time.Time
	ModifiedTo   *time.Time
	MinSize      *int64
	MaxSize      *int64
	Tier         string
	OnlyStarred  bool
	Sort         FileSearchSort
	Order        FileSearchOrder
}

type FileSearchParams struct {
	Query       string
	ParentID    int
	IsRecursive bool
	Page        int
	PageSize    int
	Filter      FileSearchFilter
}

type FileSearchQuery struct {
	Query     string
	Scope     FileSearchScope
	ScopePath string
	Filter    FileSearchFilter
	Page      int
	PageSize  int
}
