package documenttext

import "time"

const (
	StateIndexed  = "indexed"
	StatePartial  = "partial"
	StateNameOnly = "name_only"
	StateQueued   = "queued"
	StateFailed   = "failed"
)

const FailureReasonNoText = "no_text"

type DocumentTextModel struct {
	FileID          int
	ExtractedText   string
	TextLength      int
	Truncated       bool
	SourceUpdatedAt time.Time
	ErrorCode       string
	State           string
	IndexedBytes    int64
	FailureReason   string
}

type IndexStateModel struct {
	State         string
	IndexedBytes  int64
	FailureReason string
}

type PendingDocument struct {
	FileID    int
	Path      string
	Format    string
	Size      int64
	UpdatedAt time.Time
}

type DocumentMatchModel struct {
	FileID          int
	Name            string
	Path            string
	ParentPath      string
	Format          string
	Size            int64
	UpdatedAt       time.Time
	HasContentMatch bool
	ContentWindow   string
	Page            *int
}
