package jobs

import (
	"database/sql"
	"time"

	"nas-go/api/pkg/utils"
)

type JobModel struct {
	ID              string
	Type            string
	Priority        int
	ScopeJSON       string
	Status          string
	CreatedAt       time.Time
	StartedAt       sql.NullTime
	EndedAt         sql.NullTime
	CancelRequested bool
	LastError       string
}

type StepModel struct {
	ID            string
	JobID         string
	Type          string
	Status        string
	DependsOnJSON string
	Attempts      int
	MaxAttempts   int
	LastError     string
	Progress      int
	PayloadJSON   string
	CreatedAt     time.Time
	StartedAt     sql.NullTime
	EndedAt       sql.NullTime
}

type JobFilter struct {
	Status   utils.Optional[string] `filter:"status"`
	Type     utils.Optional[string] `filter:"type"`
	Priority utils.Optional[int]    `filter:"priority"`
}
