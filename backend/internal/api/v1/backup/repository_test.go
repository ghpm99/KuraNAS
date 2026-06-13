package backup

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/backup"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestRepositoryGetSettingsDocumentFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetBackupSettingsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"setting_value"}).AddRow(`{"enabled":true}`))
	mock.ExpectRollback()

	document, found, err := repo.GetSettingsDocument()
	if err != nil || !found || document != `{"enabled":true}` {
		t.Fatalf("unexpected result: %q %v %v", document, found, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetSettingsDocumentMissing(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetBackupSettingsQuery)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, found, err := repo.GetSettingsDocument()
	if err != nil || found {
		t.Fatalf("expected clean miss, got found=%v err=%v", found, err)
	}
}

func TestRepositoryUpsertSettingsDocumentCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertBackupSettingsQuery)).
		WithArgs(`{"enabled":true}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.UpsertSettingsDocument(`{"enabled":true}`); err != nil {
		t.Fatalf("UpsertSettingsDocument: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryCountPendingFiles(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CountPendingBackupQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
	mock.ExpectRollback()

	pending, err := repo.CountPendingFiles()
	if err != nil || pending != 42 {
		t.Fatalf("unexpected result: %d %v", pending, err)
	}
}

func TestRepositoryGetLastRun(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLastBackupJobQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "created_at", "started_at", "ended_at", "last_error"}).
			AddRow(7, "completed", now, now, now, ""))
	mock.ExpectRollback()

	run, found, err := repo.GetLastRun()
	if err != nil || !found {
		t.Fatalf("unexpected miss: %v %v", found, err)
	}
	if run.JobID != 7 || run.Status != "completed" || run.StartedAt == nil || run.EndedAt == nil {
		t.Fatalf("unexpected run: %+v", run)
	}
}

func TestRepositoryGetLastRunMissing(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLastBackupJobQuery)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, found, err := repo.GetLastRun()
	if err != nil || found {
		t.Fatalf("expected clean miss, got found=%v err=%v", found, err)
	}
}

func TestRepositoryStampLastBackupCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	at := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateLastBackupByPathQuery)).
		WithArgs("/mnt/dados/a.txt", at).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.StampLastBackup("/mnt/dados/a.txt", at); err != nil {
		t.Fatalf("StampLastBackup: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
