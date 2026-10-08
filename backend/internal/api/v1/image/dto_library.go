package image

import "time"

type LibraryItemDto struct {
	FileID     int        `json:"file_id"`
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	ParentPath string     `json:"parent_path"`
	Format     string     `json:"format"`
	Size       int64      `json:"size"`
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	TakenAt    *time.Time `json:"taken_at"`
	Category   string     `json:"category"`
	Starred    bool       `json:"starred"`
	Tier       string     `json:"tier"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type LibraryPageDto struct {
	Items      []LibraryItemDto `json:"items"`
	NextCursor string           `json:"next_cursor"`
	HasNext    bool             `json:"has_next"`
	Page       int              `json:"page,omitempty"`
	PageSize   int              `json:"page_size"`
}

type LibraryCountDto struct {
	Total int `json:"total"`
}

type LibraryTimelineBucketDto struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Count int `json:"count"`
}
