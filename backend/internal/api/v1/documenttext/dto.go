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
}
