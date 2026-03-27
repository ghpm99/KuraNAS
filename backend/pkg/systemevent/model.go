package systemevent

import (
	"database/sql"
	"time"
)

const DisplayTimeLayout = "02/01/2006 15:04:05"

type EventType string

const (
	EventTypeStartup  EventType = "STARTUP"
	EventTypeShutdown EventType = "SHUTDOWN"
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
