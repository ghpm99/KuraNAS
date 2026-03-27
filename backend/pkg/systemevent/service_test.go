package systemevent

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
)

type repositorySpy struct {
	dbContext *database.DbContext
	events    []EventModel
	insertErr error
}

func (r *repositorySpy) GetDbContext() *database.DbContext {
	return r.dbContext
}

func (r *repositorySpy) Insert(tx *sql.Tx, event EventModel) error {
	if r.insertErr != nil {
		return r.insertErr
	}
	r.events = append(r.events, event)
	return nil
}

func openServiceMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestServiceRecordStartup(t *testing.T) {
	db, mock := openServiceMockDB(t)

	fixedNow := time.Date(2026, 3, 27, 10, 30, 45, 0, time.FixedZone("BRT", -3*60*60))
	repo := &repositorySpy{dbContext: database.NewDbContext(db)}
	service := &Service{
		repository: repo,
		nowFn:      func() time.Time { return fixedNow },
		hostNameFn: func() (string, error) { return "kuranas-host", nil },
		processID:  func() int { return 99 },
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	if err := service.RecordStartup(); err != nil {
		t.Fatalf("expected RecordStartup success, got %v", err)
	}

	if len(repo.events) != 1 {
		t.Fatalf("expected one recorded event, got %d", len(repo.events))
	}
	stored := repo.events[0]
	if stored.EventType != EventTypeStartup {
		t.Fatalf("expected STARTUP event type, got %s", stored.EventType)
	}
	if stored.EventTimeDisplay != "27/03/2026 10:30:45" {
		t.Fatalf("unexpected formatted event time: %s", stored.EventTimeDisplay)
	}
	if !stored.HostName.Valid || stored.HostName.String != "kuranas-host" {
		t.Fatalf("expected host name to be set")
	}
	if stored.ProcessID != 99 {
		t.Fatalf("expected process id 99, got %d", stored.ProcessID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestServiceRecordShutdownPropagatesInsertError(t *testing.T) {
	db, mock := openServiceMockDB(t)
	repo := &repositorySpy{
		dbContext: database.NewDbContext(db),
		insertErr: errors.New("insert failed"),
	}
	service := &Service{
		repository: repo,
		nowFn:      time.Now,
		hostNameFn: func() (string, error) { return "", errors.New("hostname failed") },
		processID:  func() int { return 1 },
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := service.RecordShutdown()
	if err == nil {
		t.Fatalf("expected RecordShutdown error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestServiceRecordEventFailsWhenNotConfigured(t *testing.T) {
	service := &Service{}
	if err := service.RecordStartup(); err == nil {
		t.Fatalf("expected error when service is not configured")
	}
}

func TestNewService(t *testing.T) {
	service := NewService(database.NewDbContext(&sql.DB{}))
	if service == nil || service.repository == nil || service.nowFn == nil || service.hostNameFn == nil || service.processID == nil {
		t.Fatalf("expected service fully initialized")
	}
}

func TestResolveHostNameBranches(t *testing.T) {
	withoutResolver := resolveHostName(nil)
	if withoutResolver.Valid {
		t.Fatalf("expected invalid hostname when resolver is nil")
	}

	withError := resolveHostName(func() (string, error) { return "", errors.New("hostname error") })
	if withError.Valid {
		t.Fatalf("expected invalid hostname when resolver fails")
	}

	withEmpty := resolveHostName(func() (string, error) { return "", nil })
	if withEmpty.Valid {
		t.Fatalf("expected invalid hostname when value is empty")
	}
}
