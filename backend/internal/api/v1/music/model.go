package music

import (
	"database/sql"
	"time"
)

type PlayerStateModel struct {
	ID              int
	ClientID        string
	PlaylistID      sql.NullInt64
	CurrentFileID   sql.NullInt64
	CurrentPosition float64
	Volume          float64
	Shuffle         bool
	RepeatMode      string
	UpdatedAt       time.Time
}

type PlaylistModel struct {
	ID          int
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TrackCount  int
}

type PlaylistTrackModel struct {
	ID         int
	PlaylistID int
	FileID     int
	Position   int
	AddedAt    time.Time

	// Joined fields from home_file
	FileName       string
	FilePath       string
	FileParentPath string
	FileFormat     string
	FileSize       int64
	FileUpdatedAt  time.Time
	FileCreatedAt  time.Time
	LastInteraction sql.NullTime
	LastBackup      sql.NullTime
	FileType       int
	FileCheckSum   string
	FileDeletedAt  sql.NullTime
	FileStarred    bool

	// Joined fields from audio_metadata
	MetadataID          int
	MetadataFileId      int
	MetadataPath        string
	MetadataMime        string
	MetadataLength      float64
	MetadataBitrate     int
	MetadataSampleRate  int
	MetadataChannels    int
	MetadataBitrateMode int
	MetadataEncoderInfo string
	MetadataBitDepth    int
	MetadataTitle       string
	MetadataArtist      string
	MetadataAlbum       string
	MetadataAlbumArtist string
	MetadataTrackNumber string
	MetadataGenre       string
	MetadataComposer    string
	MetadataYear        string
	MetadataRecordingDate       string
	MetadataEncoder             string
	MetadataPublisher           string
	MetadataOriginalReleaseDate string
	MetadataOriginalArtist      string
	MetadataLyricist            string
	MetadataLyrics              string
	MetadataCreatedAt           time.Time
}
