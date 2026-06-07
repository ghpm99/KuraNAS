package systemevent

import (
	"database/sql"
	"time"
)

const DisplayTimeLayout = "02/01/2006 15:04:05"

type EventType string

const (
	EventTypeStartup               EventType = "STARTUP"
	EventTypeShutdown              EventType = "SHUTDOWN"
	EventTypeWorkerPoolStarted     EventType = "WORKER_POOL_STARTED"
	EventTypeScanCompleted         EventType = "SCAN_COMPLETED"
	EventTypeJobFailed             EventType = "JOB_FAILED"
	EventTypeAIProviderUnavailable EventType = "AI_PROVIDER_UNAVAILABLE"
	EventTypeOllamaDaemonStarted   EventType = "OLLAMA_DAEMON_STARTED"
	EventTypeOllamaDaemonDown      EventType = "OLLAMA_DAEMON_UNREACHABLE"
)

type EventModel struct {
	EventTime        time.Time
	EventTimeDisplay string
	EventType        EventType
	Description      string
	Source           string
	HostName         sql.NullString
	ProcessID        int
	ExtraData        sql.NullString
}
