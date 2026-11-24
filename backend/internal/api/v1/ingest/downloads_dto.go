package ingest

import "time"

const (
	DownloadStateFilterActive   = "active"
	DownloadStateFilterFinished = "finished"
)

// DownloadDto is one item of GET /ingest/downloads. The destination is the
// storage root label plus the subfolder; no absolute server path is exposed.
type DownloadDto struct {
	ID                  int        `json:"id"`
	JobID               *int       `json:"job_id"`
	URL                 string     `json:"url"`
	Title               *string    `json:"title"`
	Preset              string     `json:"preset"`
	TargetRootID        int        `json:"target_root_id"`
	TargetRootName      string     `json:"target_root_name"`
	Subfolder           string     `json:"subfolder"`
	State               string     `json:"state"`
	Progress            int        `json:"progress"`
	SpeedBytesPerSecond *int64     `json:"speed_bytes_per_second"`
	ETASeconds          *int       `json:"eta_seconds"`
	FileID              *int       `json:"file_id"`
	Error               string     `json:"error"`
	CreatedAt           time.Time  `json:"created_at"`
	EndedAt             *time.Time `json:"ended_at"`
}
