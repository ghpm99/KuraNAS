package documenttext

import "time"

type DocumentSearchResultDto struct {
	FileID     int       `json:"file_id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ParentPath string    `json:"parent_path"`
	Format     string    `json:"format"`
	Size       int64     `json:"size"`
	UpdatedAt  time.Time `json:"updated_at"`
	Snippet    string    `json:"snippet"`
	Page       *int      `json:"page"`
}

type IndexStateDto struct {
	FileID       int    `json:"file_id"`
	State        string `json:"state"`
	IndexedBytes int64  `json:"indexed_bytes"`
	Reason       string `json:"reason"`
	Error        string `json:"error"`
}

type ReindexJobDto struct {
	JobID int `json:"job_id"`
}

type DocumentTextDto struct {
	Text      string `json:"text"`
	IsPartial bool   `json:"is_partial"`
}
