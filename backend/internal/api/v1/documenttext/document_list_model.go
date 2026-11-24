package documenttext

import (
	"time"

	"nas-go/api/pkg/utils"
)

const (
	FormatImage = "image"

	SortModifiedAt = "modified_at"
	SortName       = "name"
	SortSize       = "size"
	SortCreatedAt  = "created_at"

	OrderAscending  = "asc"
	OrderDescending = "desc"
)

type DocumentFilter struct {
	Format        string
	ModifiedFrom  *time.Time
	ModifiedTo    *time.Time
	FavoritesOnly bool
	IndexState    string
}

type DocumentListParams struct {
	Filter   DocumentFilter
	FolderID int
	Sort     string
	Order    string
	Page     int
	PageSize int
}

type DocumentListQuery struct {
	Filter     DocumentFilter
	FolderPath string
	Sort       string
	Order      string
	Page       int
	PageSize   int
}

type DocumentListItemModel struct {
	ID         int
	Name       string
	Format     string
	ParentPath string
	Size       int64
	ModifiedAt time.Time
	CreatedAt  time.Time
	Starred    bool
	IndexState string
}

type DocumentItemDto struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Format     string    `json:"format"`
	ParentPath string    `json:"parent_path"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
	CreatedAt  time.Time `json:"created_at"`
	Starred    bool      `json:"starred"`
	IndexState string    `json:"index_state"`
}

type DocumentCountDto struct {
	Total int `json:"total"`
}

type DocumentSearchQuery struct {
	TermPatterns utils.SearchTermPatterns
	Terms        []string
	Filter       DocumentFilter
	FolderPath   string
	Limit        int
	Offset       int
}

type DocumentSearchParams struct {
	Query    string
	Filter   DocumentFilter
	FolderID int
	Page     int
	PageSize int
}
