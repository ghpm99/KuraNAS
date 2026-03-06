package domain

import "time"

type JobStatus string

const (
	JobStatusQueued      JobStatus = "queued"
	JobStatusRunning     JobStatus = "running"
	JobStatusPartialFail JobStatus = "partial_fail"
	JobStatusFailed      JobStatus = "failed"
	JobStatusCompleted   JobStatus = "completed"
	JobStatusCanceled    JobStatus = "canceled"
)

type StepStatus string

const (
	StepStatusQueued    StepStatus = "queued"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusCanceled  StepStatus = "canceled"
	StepStatusSkipped   StepStatus = "skipped"
)

type JobType string

const (
	JobTypeStartupScan   JobType = "startup_scan"
	JobTypeUploadProcess JobType = "upload_process"
	JobTypeFSEvent       JobType = "fs_event"
	JobTypeReindexFolder JobType = "reindex_folder"
)

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
)

type JobPriority int

const (
	JobPriorityLow JobPriority = iota + 1
	JobPriorityNormal
	JobPriorityHigh
	JobPriorityCritical
)

type Job struct {
	ID          string       `json:"id"`
	Type        JobType      `json:"type"`
	Status      JobStatus    `json:"status"`
	Priority    JobPriority  `json:"priority"`
	Scope       ScopePayload `json:"scope"`
	Attempts    int          `json:"attempts"`
	MaxAttempts int          `json:"max_attempts"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	FinishedAt  *time.Time   `json:"finished_at,omitempty"`
}

type Step struct {
	ID          string       `json:"id"`
	JobID       string       `json:"job_id"`
	Type        StepType     `json:"type"`
	Status      StepStatus   `json:"status"`
	Scope       ScopePayload `json:"scope"`
	Order       int          `json:"order"`
	Attempts    int          `json:"attempts"`
	MaxAttempts int          `json:"max_attempts"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	FinishedAt  *time.Time   `json:"finished_at,omitempty"`
}
