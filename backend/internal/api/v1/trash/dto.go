package trash

import "time"

// TrashItemDto is the transport shape. trash_path stays server-side: clients
// only ever address items by id, never by their location inside the trash dir.
type TrashItemDto struct {
	ID           int       `json:"id"`
	OriginalPath string    `json:"original_path"`
	Size         int64     `json:"size"`
	DeletedAt    time.Time `json:"deleted_at"`
}

type RetentionDto struct {
	Days int `json:"days"`
}
