package jobs

import "time"

type JobScopeDto map[string]any

type JobProgressDto struct {
	TotalSteps     int `json:"total_steps"`
	CompletedSteps int `json:"completed_steps"`
	RunningSteps   int `json:"running_steps"`
	FailedSteps    int `json:"failed_steps"`
	SkippedSteps   int `json:"skipped_steps"`
	CanceledSteps  int `json:"canceled_steps"`
	Progress       int `json:"progress"`
}

type JobDto struct {
	ID              int            `json:"id"`
	Type            string         `json:"type"`
	Priority        string         `json:"priority"`
	Scope           JobScopeDto    `json:"scope,omitempty"`
	Status          string         `json:"status"`
	Progress        JobProgressDto `json:"progress"`
	CreatedAt       time.Time      `json:"created_at"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	EndedAt         *time.Time     `json:"ended_at,omitempty"`
	CancelRequested bool           `json:"cancel_requested"`
	LastError       string         `json:"last_error,omitempty"`
}

type StepDto struct {
	ID          int        `json:"id"`
	JobID       int        `json:"job_id"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	DependsOn   []int      `json:"depends_on,omitempty"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	LastError   string     `json:"last_error,omitempty"`
	Progress    int        `json:"progress"`
	Payload     any        `json:"payload,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
}
