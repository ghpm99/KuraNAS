package watchfolders

import "time"

type WatchFolderDto struct {
	ID         int        `json:"id"`
	Path       string     `json:"path"`
	Label      string     `json:"label,omitempty"`
	Enabled    bool       `json:"enabled"`
	LastScanAt *time.Time `json:"last_scan_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateWatchFolderDto struct {
	Path  string `json:"path" binding:"required"`
	Label string `json:"label"`
}

type UpdateWatchFolderDto struct {
	Path    *string `json:"path,omitempty"`
	Label   *string `json:"label,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

func (m *WatchFolderModel) ToDto() WatchFolderDto {
	return WatchFolderDto{
		ID:         m.ID,
		Path:       m.Path,
		Label:      m.Label,
		Enabled:    m.Enabled,
		LastScanAt: m.LastScanAt,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func (d *WatchFolderDto) ToModel() WatchFolderModel {
	return WatchFolderModel{
		ID:         d.ID,
		Path:       d.Path,
		Label:      d.Label,
		Enabled:    d.Enabled,
		LastScanAt: d.LastScanAt,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}
