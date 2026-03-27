package systemevent

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	"os"
	"time"
)

const (
	defaultSource       = "backend"
	startupDescription  = "KuraNAS system startup"
	shutdownDescription = "KuraNAS system shutdown"
)

type Service struct {
	repository RepositoryInterface
	nowFn      func() time.Time
	hostNameFn func() (string, error)
	processID  func() int
}

func NewService(dbContext *database.DbContext) *Service {
	return &Service{
		repository: NewRepository(dbContext),
		nowFn:      time.Now,
		hostNameFn: os.Hostname,
		processID:  os.Getpid,
	}
}

func (s *Service) RecordStartup() error {
	return s.recordEvent(EventTypeStartup, startupDescription)
}

func (s *Service) RecordShutdown() error {
	return s.recordEvent(EventTypeShutdown, shutdownDescription)
}

func (s *Service) recordEvent(eventType EventType, description string) error {
	if s == nil || s.repository == nil || s.repository.GetDbContext() == nil {
		return fmt.Errorf("system event service is not configured")
	}

	now := s.nowFn()
	event := EventModel{
		EventTime:        now,
		EventTimeDisplay: now.Format(DisplayTimeLayout),
		EventType:        eventType,
		Description:      description,
		Source:           defaultSource,
		HostName:         resolveHostName(s.hostNameFn),
		ProcessID:        s.processID(),
	}

	err := s.repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return s.repository.Insert(tx, event)
	})
	if err != nil {
		return fmt.Errorf("record system event: %w", err)
	}

	return nil
}

func resolveHostName(hostNameFn func() (string, error)) sql.NullString {
	if hostNameFn == nil {
		return sql.NullString{Valid: false}
	}

	hostName, err := hostNameFn()
	if err != nil || hostName == "" {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: hostName, Valid: true}
}
