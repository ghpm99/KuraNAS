package documenttext

import "time"

type DocumentTextModel struct {
	FileID          int
	ExtractedText   string
	TextLength      int
	Truncated       bool
	SourceUpdatedAt time.Time
	ErrorCode       string
}

type PendingDocument struct {
	FileID    int
	Path      string
	Format    string
	Size      int64
	UpdatedAt time.Time
}

type DocumentMatchModel struct {
	FileID        int
	Name          string
	Path          string
	ParentPath    string
	Format        string
	Size          int64
	UpdatedAt     time.Time
	ExtractedText string
}
