package storageroots

import "time"

type StorageRootDto struct {
	ID        int       `json:"id"`
	Path      string    `json:"path"`
	Label     string    `json:"label"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStorageRootDto struct {
	Path    string `json:"path"`
	Label   string `json:"label"`
	Enabled *bool  `json:"enabled"`
}

// UpdateStorageRootDto changes label/enabled; the path is immutable (delete
// and re-add to move a root).
type UpdateStorageRootDto struct {
	Label   string `json:"label"`
	Enabled *bool  `json:"enabled"`
}
