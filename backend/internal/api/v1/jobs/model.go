package jobs

import "time"

type JobModel struct {
	ID              int
	Type            string
	Priority        string
	Scope           []byte
	Status          string
	CreatedAt       time.Time
	StartedAt       *time.Time
	EndedAt         *time.Time
	CancelRequested bool
	LastError       string
}

type StepModel struct {
	ID          int
	JobID       int
	Type        string
	Status      string
	DependsOn   []byte
	Attempts    int
	MaxAttempts int
	LastError   string
	Progress    int
	Payload     []byte
	CreatedAt   time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
}
