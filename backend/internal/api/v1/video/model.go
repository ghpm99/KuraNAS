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
	ID             int
	Type           string
	SourcePath     string
	Name           string
	IsHidden       bool
	IsAuto         bool
	GroupMode      string
	Classification string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastPlayedAt   sql.NullTime
	ItemCount      int
	CoverVideoID   sql.NullInt64
}

type VideoPlaylistItemModel struct {
	ID         int
	PlaylistID int
	VideoID    int
	OrderIndex int
	SourceKind string
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

type VideoBehaviorEventModel struct {
	ID         int
	ClientID   string
	VideoID    int
	PlaylistID int
	EventType  string
	Position   float64
	Duration   float64
	WatchedPct float64
	CreatedAt  time.Time
}

// VideoWithMetadataModel extends VideoFileModel with optional video_metadata fields.
type VideoWithMetadataModel struct {
	VideoFileModel
	// Nullable metadata fields (LEFT JOIN)
	MetaDuration        sql.NullString
	MetaWidth           sql.NullInt64
	MetaHeight          sql.NullInt64
	MetaFrameRate       sql.NullFloat64
	MetaCodecName       sql.NullString
	MetaAspectRatio     sql.NullString
	MetaAudioChannels   sql.NullInt64
	MetaAudioCodec      sql.NullString
	MetaAudioSampleRate sql.NullString
}
