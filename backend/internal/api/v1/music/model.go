package music

import (
	"database/sql"
	"time"
)

// AudioMetadataModel is the DB shape of audio-specific metadata stored in the
// audio_metadata complement table. It mirrors files.AudioMetadataModel but lives
// here so the music extension owns its own complement table and files never
// imports music.
type AudioMetadataModel struct {
	ID                  int
	FileId              int
	Path                string
	Mime                string  `json:"mime"`
	Length              float64 `json:"length"`
	Bitrate             int     `json:"bitrate"`
	SampleRate          int     `json:"sample_rate"`
	Channels            int     `json:"channels"`
	BitrateMode         int     `json:"bitrate_mode"`
	EncoderInfo         string  `json:"encoder_info"`
	BitDepth            int     `json:"bit_depth"`
	Title               string  `json:"title"`
	Artist              string  `json:"artist"`
	Album               string  `json:"album"`
	AlbumArtist         string  `json:"album_artist"`
	TrackNumber         string  `json:"track_number"`
	Genre               string  `json:"genre"`
	Composer            string  `json:"composer"`
	Year                string  `json:"year"`
	RecordingDate       string  `json:"recording_date"`
	Encoder             string  `json:"encoder"`
	Publisher           string  `json:"publisher"`
	OriginalReleaseDate string  `json:"original_release_date"`
	OriginalArtist      string  `json:"original_artist"`
	Lyricist            string  `json:"lyricist"`
	Lyrics              string  `json:"lyrics"`
	CreatedAt           time.Time
}

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
	ID            int
	Name          string
	Description   string
	IsSystem      bool
	IsAIGenerated bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	TrackCount    int
}

// ArtistClusterModel is the persisted assignment of one artist to one
// AI-generated playlist category. It is the source of truth the worker rebuilds
// the AI playlists from, and lets clustering stay incremental (only new artists
// are sent to the model).
type ArtistClusterModel struct {
	ArtistKey   string
	Artist      string
	ClusterName string
}

type PlaylistTrackModel struct {
	ID         int
	PlaylistID int
	FileID     int
	Position   int
	AddedAt    time.Time

	// Joined fields from home_file
	FileName        string
	FilePath        string
	FileParentPath  string
	FileFormat      string
	FileSize        int64
	FileUpdatedAt   time.Time
	FileCreatedAt   time.Time
	LastInteraction sql.NullTime
	LastBackup      sql.NullTime
	FileType        int
	FileCheckSum    string
	FileDeletedAt   sql.NullTime
	FileStarred     bool

	// Joined fields from audio_metadata
	MetadataID                  int
	MetadataFileId              int
	MetadataPath                string
	MetadataMime                string
	MetadataLength              float64
	MetadataBitrate             int
	MetadataSampleRate          int
	MetadataChannels            int
	MetadataBitrateMode         int
	MetadataEncoderInfo         string
	MetadataBitDepth            int
	MetadataTitle               string
	MetadataArtist              string
	MetadataAlbum               string
	MetadataAlbumArtist         string
	MetadataTrackNumber         string
	MetadataGenre               string
	MetadataComposer            string
	MetadataYear                string
	MetadataRecordingDate       string
	MetadataEncoder             string
	MetadataPublisher           string
	MetadataOriginalReleaseDate string
	MetadataOriginalArtist      string
	MetadataLyricist            string
	MetadataLyrics              string
	MetadataCreatedAt           time.Time
}
