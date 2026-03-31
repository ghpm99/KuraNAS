package captures

import "time"

type CaptureModel struct {
	ID        int
	Name      string
	FileName  string
	FilePath  string
	MediaType string
	MimeType  string
	Size      int64
	CreatedAt time.Time
}
