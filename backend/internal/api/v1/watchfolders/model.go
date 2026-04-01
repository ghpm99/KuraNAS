package watchfolders

import "time"

type WatchFolderModel struct {
	ID         int
	Path       string
	Label      string
	Enabled    bool
	LastScanAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
