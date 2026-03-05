package logger

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"nas-go/api/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
)

type loggerRepoMock struct {
	dbCtx        *database.DbContext
	createLogFn  func(tx *sql.Tx, log LoggerModel) (LoggerModel, error)
	getLogByIDFn func(id int) (LoggerModel, error)
	getLogsFn    func(page, pageSize int) ([]LoggerModel, error)
	updateLogFn  func(tx *sql.Tx, log LoggerModel) error
}

func (m *loggerRepoMock) GetDbContext() *database.DbContext { return m.dbCtx }
func (m *loggerRepoMock) CreateLog(tx *sql.Tx, log LoggerModel) (LoggerModel, error) {
	return m.createLogFn(tx, log)
}
func (m *loggerRepoMock) GetLogByID(id int) (LoggerModel, error) { return m.getLogByIDFn(id) }
func (m *loggerRepoMock) GetLogs(page, pageSize int) ([]LoggerModel, error) {
	return m.getLogsFn(page, pageSize)
}
func (m *loggerRepoMock) UpdateLog(tx *sql.Tx, log LoggerModel) error { return m.updateLogFn(tx, log) }

func newLoggerServiceWithMock(t *testing.T, repo *loggerRepoMock) (sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	repo.dbCtx = database.NewDbContext(db)
	return mock, db
}

func TestLoggerServiceCreateLogAndGetters(t *testing.T) {
	repo := &loggerRepoMock{
		createLogFn: func(tx *sql.Tx, log LoggerModel) (LoggerModel, error) {
			log.ID = 10
			return log, nil
		},
		getLogByIDFn: func(id int) (LoggerModel, error) {
			return LoggerModel{ID: id}, nil
		},
		getLogsFn: func(page, pageSize int) ([]LoggerModel, error) {
			return []LoggerModel{{ID: 1}}, nil
		},
		updateLogFn: func(tx *sql.Tx, log LoggerModel) error { return nil },
	}
	mock, db := newLoggerServiceWithMock(t, repo)
	defer db.Close()
	svc := NewLoggerService(repo)

	mock.ExpectBegin()
	mock.ExpectCommit()
	created, err := svc.CreateLog(LoggerModel{
		Name:   "op",
		Level:  LogLevelInfo,
		Status: LogStatusPending,
	}, map[string]any{"x": 1})
	if err != nil {
		t.Fatalf("CreateLog returned error: %v", err)
	}
	if created.ID != 10 {
		t.Fatalf("expected ID 10, got %d", created.ID)
	}

	got, err := svc.GetLogByID(7)
	if err != nil || got.ID != 7 {
		t.Fatalf("GetLogByID failed, got=%+v err=%v", got, err)
	}

	list, err := svc.GetLogs(1, 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("GetLogs failed, len=%d err=%v", len(list), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestLoggerServiceCreateAndUpdateErrors(t *testing.T) {
	repo := &loggerRepoMock{
		createLogFn: func(tx *sql.Tx, log LoggerModel) (LoggerModel, error) {
			return LoggerModel{}, errors.New("create failed")
		},
		getLogByIDFn: func(id int) (LoggerModel, error) { return LoggerModel{}, nil },
		getLogsFn:    func(page, pageSize int) ([]LoggerModel, error) { return nil, nil },
		updateLogFn: func(tx *sql.Tx, log LoggerModel) error {
			return errors.New("update failed")
		},
	}
	mock, db := newLoggerServiceWithMock(t, repo)
	defer db.Close()
	svc := NewLoggerService(repo)

	mock.ExpectBegin()
	mock.ExpectRollback()
	_, err := svc.CreateLog(LoggerModel{
		Name:   "op",
		Level:  LogLevelInfo,
		Status: LogStatusPending,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "error creating log") {
		t.Fatalf("expected wrapped create error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectRollback()
	if err := svc.UpdateLog(LoggerModel{ID: 1}); err == nil || !strings.Contains(err.Error(), "error updating log") {
		t.Fatalf("expected wrapped update error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestLoggerServiceCompleteHelpers(t *testing.T) {
	repo := &loggerRepoMock{
		createLogFn:  func(tx *sql.Tx, log LoggerModel) (LoggerModel, error) { return log, nil },
		getLogByIDFn: func(id int) (LoggerModel, error) { return LoggerModel{}, nil },
		getLogsFn:    func(page, pageSize int) ([]LoggerModel, error) { return nil, nil },
		updateLogFn:  func(tx *sql.Tx, log LoggerModel) error { return nil },
	}
	mock, db := newLoggerServiceWithMock(t, repo)
	defer db.Close()
	svc := NewLoggerService(repo)

	mock.ExpectBegin()
	mock.ExpectCommit()
	if err := svc.CompleteWithSuccessLog(LoggerModel{ID: 1}); err != nil {
		t.Fatalf("CompleteWithSuccessLog returned error: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectCommit()
	if err := svc.CompleteWithErrorLog(LoggerModel{ID: 2}, errors.New("boom")); err != nil {
		t.Fatalf("CompleteWithErrorLog returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
