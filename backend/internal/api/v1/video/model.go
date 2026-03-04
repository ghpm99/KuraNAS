package video

import (
	"database/sql"
	"time"
)

type ContextType string

const (
	ContextFolder ContextType = "folder"
)

type VideoFileModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Format     string
	Size       int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type VideoPlaylistModel struct {
	ID           int
	Type         string
	SourcePath   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastPlayedAt sql.NullTime
}

type VideoPlaylistItemModel struct {
	ID         int
	PlaylistID int
	VideoID    int
	OrderIndex int
	Video      VideoFileModel
}

type VideoPlaybackStateModel struct {
	ID          int
	ClientID    string
	PlaylistID  sql.NullInt64
	VideoID     sql.NullInt64
	CurrentTime float64
	Duration    float64
	IsPaused    bool
	Completed   bool
	LastUpdate  time.Time
}
