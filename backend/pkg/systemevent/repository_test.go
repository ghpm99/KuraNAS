package systemevent

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/systemevent"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := NewRepository(database.NewDbContext(db))

	now := time.Now()
	event := EventModel{
		EventTime:        now,
		EventTimeDisplay: now.Format(DisplayTimeLayout),
		EventType:        EventTypeStartup,
		Description:      "startup",
		Source:           "backend",
		HostName:         sql.NullString{String: "host", Valid: true},
		ProcessID:        42,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertSystemEventQuery)).
		WithArgs(
			event.EventTime,
			event.EventTimeDisplay,
			event.EventType,
			event.Description,
			event.Source,
			event.HostName,
			event.ProcessID,
			event.ExtraData,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repository.Insert(tx, event)
	})
	if err != nil {
		t.Fatalf("expected insert success, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestRepositoryInsertFailsWhenTxIsNil(t *testing.T) {
	repository := NewRepository(nil)
	err := repository.Insert(nil, EventModel{})
	if err == nil {
		t.Fatalf("expected error when tx is nil")
	}
}

func TestRepositoryInsertPropagatesExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := NewRepository(database.NewDbContext(db))

	now := time.Now()
	event := EventModel{
		EventTime:        now,
		EventTimeDisplay: now.Format(DisplayTimeLayout),
		EventType:        EventTypeShutdown,
		Description:      "shutdown",
		Source:           "backend",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertSystemEventQuery)).
		WithArgs(
			event.EventTime,
			event.EventTimeDisplay,
			event.EventType,
			event.Description,
			event.Source,
			event.HostName,
			event.ProcessID,
			event.ExtraData,
		).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	err = repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repository.Insert(tx, event)
	})
	if err == nil {
		t.Fatalf("expected insert error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
