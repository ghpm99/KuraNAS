package video

import "time"

type VideoFileDto struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	ParentPath string `json:"parent_path"`
	Format     string `json:"format"`
	Size       int64  `json:"size"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type VideoPlaylistDto struct {
	ID             int                    `json:"id"`
	Type           string                 `json:"type"`
	SourcePath     string                 `json:"source_path"`
	Name           string                 `json:"name"`
	IsHidden       bool                   `json:"is_hidden"`
	IsAuto         bool                   `json:"is_auto"`
	GroupMode      string                 `json:"group_mode"`
	Classification string                 `json:"classification"`
	ItemCount      int                    `json:"item_count"`
	CoverVideoID   *int                   `json:"cover_video_id"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	LastPlayedAt   *time.Time             `json:"last_played_at"`
	Items          []VideoPlaylistItemDto `json:"items"`
}

type VideoPlaylistItemDto struct {
	ID         int          `json:"id"`
	OrderIndex int          `json:"order_index"`
	SourceKind string       `json:"source_kind"`
	Video      VideoFileDto `json:"video"`
	Status     string       `json:"status"`
}

type VideoPlaybackStateDto struct {
	ID          int       `json:"id"`
	ClientID    string    `json:"client_id"`
	PlaylistID  *int      `json:"playlist_id"`
	VideoID     *int      `json:"video_id"`
	CurrentTime float64   `json:"current_time"`
	Duration    float64   `json:"duration"`
	IsPaused    bool      `json:"is_paused"`
	Completed   bool      `json:"completed"`
	LastUpdate  time.Time `json:"last_update"`
}

type PlaybackSessionDto struct {
	Playlist      VideoPlaylistDto      `json:"playlist"`
	PlaybackState VideoPlaybackStateDto `json:"playback_state"`
}

type VideoCatalogItemDto struct {
	Video       VideoFileDto `json:"video"`
	Status      string       `json:"status"`
	ProgressPct float64      `json:"progress_pct"`
}

type VideoCatalogSectionDto struct {
	Key   string                `json:"key"`
	Title string                `json:"title"`
	Items []VideoCatalogItemDto `json:"items"`
}

type VideoHomeCatalogDto struct {
	Sections []VideoCatalogSectionDto `json:"sections"`
}

type StartPlaybackRequest struct {
	VideoID    int  `json:"video_id" binding:"required"`
	PlaylistID *int `json:"playlist_id"`
}

type UpdatePlaybackStateRequest struct {
	PlaylistID  *int     `json:"playlist_id"`
	VideoID     *int     `json:"video_id"`
	CurrentTime *float64 `json:"current_time"`
	Duration    *float64 `json:"duration"`
	IsPaused    *bool    `json:"is_paused"`
	Completed   *bool    `json:"completed"`
}

type SetPlaylistHiddenRequest struct {
	Hidden bool `json:"hidden"`
}

type AddPlaylistVideoRequest struct {
	VideoID int `json:"video_id" binding:"required"`
}

type UpdatePlaylistRequest struct {
	Name string `json:"name" binding:"required"`
}

type ReorderPlaylistItemRequest struct {
	VideoID    int `json:"video_id" binding:"required"`
	OrderIndex int `json:"order_index" binding:"required"`
}

type ReorderPlaylistRequest struct {
	Items []ReorderPlaylistItemRequest `json:"items" binding:"required"`
}

func (m *VideoFileModel) ToDto() VideoFileDto {
	return VideoFileDto{
		ID:         m.ID,
		Name:       m.Name,
		Path:       m.Path,
		ParentPath: m.ParentPath,
		Format:     m.Format,
		Size:       m.Size,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
	}
}

func (m *VideoPlaylistModel) ToDto(items []VideoPlaylistItemDto) VideoPlaylistDto {
	dto := VideoPlaylistDto{
		ID:             m.ID,
		Type:           m.Type,
		SourcePath:     m.SourcePath,
		Name:           m.Name,
		IsHidden:       m.IsHidden,
		IsAuto:         m.IsAuto,
		GroupMode:      m.GroupMode,
		Classification: m.Classification,
		ItemCount:      m.ItemCount,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		Items:          items,
	}
	if m.LastPlayedAt.Valid {
		t := m.LastPlayedAt.Time
		dto.LastPlayedAt = &t
	}
	if m.CoverVideoID.Valid {
		v := int(m.CoverVideoID.Int64)
		dto.CoverVideoID = &v
	}
	return dto
}

func (m *VideoPlaylistItemModel) ToDto(status string) VideoPlaylistItemDto {
	return VideoPlaylistItemDto{
		ID:         m.ID,
		OrderIndex: m.OrderIndex,
		SourceKind: m.SourceKind,
		Video:      m.Video.ToDto(),
		Status:     status,
	}
}

func (m *VideoPlaybackStateModel) ToDto() VideoPlaybackStateDto {
	dto := VideoPlaybackStateDto{
		ID:          m.ID,
		ClientID:    m.ClientID,
		CurrentTime: m.CurrentTime,
		Duration:    m.Duration,
		IsPaused:    m.IsPaused,
		Completed:   m.Completed,
		LastUpdate:  m.LastUpdate,
	}
	if m.PlaylistID.Valid {
		v := int(m.PlaylistID.Int64)
		dto.PlaylistID = &v
	}
	if m.VideoID.Valid {
		v := int(m.VideoID.Int64)
		dto.VideoID = &v
	}
	return dto
}
