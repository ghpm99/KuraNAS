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
	ID           int                    `json:"id"`
	Type         string                 `json:"type"`
	SourcePath   string                 `json:"source_path"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastPlayedAt *time.Time             `json:"last_played_at"`
	Items        []VideoPlaylistItemDto `json:"items"`
}

type VideoPlaylistItemDto struct {
	ID         int          `json:"id"`
	OrderIndex int          `json:"order_index"`
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
	VideoID int `json:"video_id" binding:"required"`
}

type UpdatePlaybackStateRequest struct {
	PlaylistID  *int     `json:"playlist_id"`
	VideoID     *int     `json:"video_id"`
	CurrentTime *float64 `json:"current_time"`
	Duration    *float64 `json:"duration"`
	IsPaused    *bool    `json:"is_paused"`
	Completed   *bool    `json:"completed"`
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
		ID:         m.ID,
		Type:       m.Type,
		SourcePath: m.SourcePath,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		Items:      items,
	}
	if m.LastPlayedAt.Valid {
		t := m.LastPlayedAt.Time
		dto.LastPlayedAt = &t
	}
	return dto
}

func (m *VideoPlaylistItemModel) ToDto(status string) VideoPlaylistItemDto {
	return VideoPlaylistItemDto{
		ID:         m.ID,
		OrderIndex: m.OrderIndex,
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
