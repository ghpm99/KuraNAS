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

// CaptureModel is the source of truth for a capture's rich metadata. The scalar
// semantic columns are filled during promotion by parsing RawMetadata; the raw
// payload is preserved verbatim so future jobs can re-derive fields without the
// retired metadata.json sidecar. FileID links the promoted recording to its
// home_file row (nil until promotion pre-registers it).
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
}
