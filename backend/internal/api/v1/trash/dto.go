package trash

import "time"

// TrashItemDto is the transport shape. trash_path stays server-side: clients
// only ever address items by id, never by their location inside the trash dir.
type TrashItemDto struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	OriginalPath string    `json:"original_path"`
	Size         int64     `json:"size"`
	DeletedAt    time.Time `json:"deleted_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsDirectory  bool      `json:"is_directory"`
}

type RetentionDto struct {
	Days int `json:"days"`
}
