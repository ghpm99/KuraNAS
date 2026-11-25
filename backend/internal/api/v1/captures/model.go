package captures

import (
	"encoding/json"
	"time"
)

// CaptureStatus tracks where a capture is in the upload → promotion lifecycle.
type CaptureStatus string

const (
	CaptureStatusUploaded  CaptureStatus = "uploaded"
	CaptureStatusPromoting CaptureStatus = "promoting"
	CaptureStatusPromoted  CaptureStatus = "promoted"
	CaptureStatusFailed    CaptureStatus = "failed"
)

// CaptureModel is the source of truth for a capture and its metadata. RawMetadata
// keeps the client payload verbatim; the semantic columns are derived from it
// during promotion. FileID is nil until promotion pre-registers the home_file row.
type CaptureModel struct {
	ID         int
	Name       string
	FileName   string
	FilePath   string
	MediaType  string
	MimeType   string
	Size       int64
	EpisodeKey string
	CreatedAt  time.Time

	FileID        *int
	Status        CaptureStatus
	Title         string
	EpisodeTitle  string
	Season        *int
	Episode       *int
	Description   string
	ReleaseYear   *int
	Genres        []string
	Cast          []string
	Directors     []string
	Studio        string
	ContentRating string
	Platform      string
	SourceURL     string
	ThumbnailURL  string
	ContentType   string
	RawMetadata   json.RawMessage
	FailureReason string
}
