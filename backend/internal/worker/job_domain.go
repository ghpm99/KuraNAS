package worker

import "time"

type JobType string

const (
	JobTypeStartupScan   JobType = "startup_scan"
	JobTypeUploadProcess JobType = "upload_process"
	JobTypeFSEvent       JobType = "fs_event"
	JobTypeReindexFolder JobType = "reindex_folder"
	JobTypeTakeoutImport JobType = "takeout_import"
)

func (t JobType) IsValid() bool {
	switch t {
	case JobTypeStartupScan, JobTypeUploadProcess, JobTypeFSEvent, JobTypeReindexFolder, JobTypeTakeoutImport:
		return true
	default:
		return false
	}
}

type StepType string

const (
	StepTypeScanFilesystem StepType = "scan_filesystem"
	StepTypeDiffAgainstDB  StepType = "diff_against_db"
	StepTypeMetadata       StepType = "metadata"
	StepTypeChecksum       StepType = "checksum"
	StepTypePersist        StepType = "persist"
	StepTypeThumbnail      StepType = "thumbnail"
	StepTypePlaylistIndex  StepType = "playlist_index"
	StepTypeMarkDeleted    StepType = "mark_deleted"
	StepTypeTakeoutExtract StepType = "takeout_extract"
)

func (t StepType) IsValid() bool {
	switch t {
	case StepTypeScanFilesystem, StepTypeDiffAgainstDB, StepTypeMetadata, StepTypeChecksum, StepTypePersist, StepTypeThumbnail, StepTypePlaylistIndex, StepTypeMarkDeleted, StepTypeTakeoutExtract:
		return true
	default:
		return false
	}
}

type JobPriority string

const (
	JobPriorityLow    JobPriority = "low"
	JobPriorityNormal JobPriority = "normal"
	JobPriorityHigh   JobPriority = "high"
)

func (p JobPriority) IsValid() bool {
	switch p {
	case JobPriorityLow, JobPriorityNormal, JobPriorityHigh:
		return true
	default:
		return false
	}
}

func (p JobPriority) Weight() int {
	switch p {
	case JobPriorityHigh:
		return 3
	case JobPriorityNormal:
		return 2
	case JobPriorityLow:
		return 1
	default:
		return 0
	}
}

type JobStatus string

const (
	JobStatusQueued      JobStatus = "queued"
	JobStatusRunning     JobStatus = "running"
	JobStatusPartialFail JobStatus = "partial_fail"
	JobStatusFailed      JobStatus = "failed"
	JobStatusCompleted   JobStatus = "completed"
	JobStatusCanceled    JobStatus = "canceled"
)

func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusQueued, JobStatusRunning, JobStatusPartialFail, JobStatusFailed, JobStatusCompleted, JobStatusCanceled:
		return true
	default:
		return false
	}
}

type StepStatus string

const (
	StepStatusQueued    StepStatus = "queued"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusCanceled  StepStatus = "canceled"
	StepStatusSkipped   StepStatus = "skipped"
)

func (s StepStatus) IsValid() bool {
	switch s {
	case StepStatusQueued, StepStatusRunning, StepStatusCompleted, StepStatusFailed, StepStatusCanceled, StepStatusSkipped:
		return true
	default:
		return false
	}
}

type JobScope struct {
	FileID *int   `json:"file_id,omitempty"`
	Path   string `json:"path,omitempty"`
	Root   string `json:"root,omitempty"`
}

func (s JobScope) IsEmpty() bool {
	return s.FileID == nil && s.Path == "" && s.Root == ""
}

type Job struct {
	ID              int         `json:"id"`
	Type            JobType     `json:"type"`
	Priority        JobPriority `json:"priority"`
	Scope           JobScope    `json:"scope"`
	Status          JobStatus   `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
	StartedAt       *time.Time  `json:"started_at,omitempty"`
	EndedAt         *time.Time  `json:"ended_at,omitempty"`
	CancelRequested bool        `json:"cancel_requested"`
	LastError       string      `json:"last_error,omitempty"`
}

type Step struct {
	ID          int        `json:"id"`
	JobID       int        `json:"job_id"`
	Type        StepType   `json:"type"`
	Status      StepStatus `json:"status"`
	DependsOn   []int      `json:"depends_on,omitempty"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	Progress    int        `json:"progress"`
	LastError   string     `json:"last_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
}

func (s Step) IsTerminal() bool {
	switch s.Status {
	case StepStatusCompleted, StepStatusFailed, StepStatusCanceled, StepStatusSkipped:
		return true
	default:
		return false
	}
}
