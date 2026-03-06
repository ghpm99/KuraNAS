package jobs

import "time"

type JobProgressSummaryDto struct {
	Progress       int `json:"progress"`
	TotalSteps     int `json:"total_steps"`
	QueuedSteps    int `json:"queued_steps"`
	RunningSteps   int `json:"running_steps"`
	CompletedSteps int `json:"completed_steps"`
	FailedSteps    int `json:"failed_steps"`
	CanceledSteps  int `json:"canceled_steps"`
	SkippedSteps   int `json:"skipped_steps"`
}

type JobSummaryDto struct {
	ID              string                `json:"id"`
	Type            string                `json:"type"`
	Priority        int                   `json:"priority"`
	ParentJobID     *string               `json:"parent_job_id,omitempty"`
	ScopeJSON       string                `json:"scope_json"`
	Status          string                `json:"status"`
	CreatedAt       time.Time             `json:"created_at"`
	StartedAt       *time.Time            `json:"started_at,omitempty"`
	EndedAt         *time.Time            `json:"ended_at,omitempty"`
	CancelRequested bool                  `json:"cancel_requested"`
	LastError       string                `json:"last_error"`
	Progress        JobProgressSummaryDto `json:"progress_summary"`
}

type StepDto struct {
	ID            string     `json:"id"`
	JobID         string     `json:"job_id"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	DependsOnJSON string     `json:"depends_on_json"`
	Attempts      int        `json:"attempts"`
	MaxAttempts   int        `json:"max_attempts"`
	LastError     string     `json:"last_error"`
	Progress      int        `json:"progress"`
	PayloadJSON   string     `json:"payload_json"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
}
