package logger

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/log"

	"github.com/DATA-DOG/go-sqlmock"
)

func newLoggerRepoWithMock(t *testing.T) (*LoggerRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	repo := NewLoggerRepository(database.NewDbContext(db))
	return repo, mock, db
}

func sampleLogModel() LoggerModel {
	now := time.Now()
	return LoggerModel{
		ID:          1,
		Name:        "test",
		Description: "desc",
		Level:       LogLevelInfo,
		IPAddress:   "127.0.0.1",
		StartTime:   now,
		EndTime:     sql.NullTime{Time: now, Valid: true},
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sql.NullTime{},
		Status:      LogStatusPending,
		ExtraData:   sql.NullString{String: `{"ok":true}`, Valid: true},
	}
}

func TestLoggerRepositoryCreateLog(t *testing.T) {
	repo, mock, db := newLoggerRepoWithMock(t)
	defer db.Close()

	model := sampleLogModel()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertLogQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, err := repo.CreateLog(tx, model)
		if err != nil {
			return err
		}
		if created.ID != 99 {
			t.Fatalf("expected returned id 99, got %d", created.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLoggerRepositoryGetLogByIDAndGetLogs(t *testing.T) {
	repo, mock, db := newLoggerRepoWithMock(t)
	defer db.Close()

	row := sqlmock.NewRows([]string{
		"id", "name", "description", "level", "ip_address",
		"start_time", "end_time", "created_at", "updated_at",
		"deleted_at", "status", "extra_data",
	}).AddRow(
		1, "name", "desc", string(LogLevelInfo), "127.0.0.1",
		time.Now(), nil, time.Now(), time.Now(),
		nil, string(LogStatusPending), `{"data":{"a":1}}`,
	)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLogByIDQuery)).
		WithArgs(1).
		WillReturnRows(row)

	logResult, err := repo.GetLogByID(1)
	if err != nil {
		t.Fatalf("GetLogByID returned error: %v", err)
	}
	if logResult.ID != 1 {
		t.Fatalf("expected log id 1, got %d", logResult.ID)
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "level", "ip_address",
		"start_time", "end_time", "created_at", "updated_at",
		"deleted_at", "status", "extra_data",
	}).AddRow(
		1, "name", "desc", string(LogLevelInfo), "127.0.0.1",
		time.Now(), nil, time.Now(), time.Now(),
		nil, string(LogStatusPending), `{"data":{"a":1}}`,
	)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLogsQuery)).
		WithArgs(10, 0).
		WillReturnRows(rows)

	logs, err := repo.GetLogs(1, 10)
	if err != nil {
		t.Fatalf("GetLogs returned error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected one row, got %d", len(logs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLoggerRepositoryUpdateLog(t *testing.T) {
	repo, mock, db := newLoggerRepoWithMock(t)
	defer db.Close()

	model := sampleLogModel()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateLogQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.UpdateLog(tx, model)
	})
	if err != nil {
		t.Fatalf("UpdateLog returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
