package vacuum

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/vacuum"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRepositoryWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(database.NewDbContext(db)), mock
}

func TestRepositoryGetSettingsDocumentFound(t *testing.T) {
	repository, mock := newRepositoryWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetSettingsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"setting_value"}).AddRow(`{"device_id":"x"}`))
	mock.ExpectRollback()

	document, isFound, err := repository.GetSettingsDocument()
	if err != nil || !isFound || document != `{"device_id":"x"}` {
		t.Fatalf("unexpected result: %q %v %v", document, isFound, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetSettingsDocumentMissing(t *testing.T) {
	repository, mock := newRepositoryWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetSettingsQuery)).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, isFound, err := repository.GetSettingsDocument()
	if err != nil || isFound {
		t.Fatalf("expected a clean miss, got isFound=%v err=%v", isFound, err)
	}
}

func TestRepositoryGetSettingsDocumentFailure(t *testing.T) {
	repository, mock := newRepositoryWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetSettingsQuery)).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	if _, _, err := repository.GetSettingsDocument(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestRepositoryUpsertSettingsDocumentCommits(t *testing.T) {
	repository, mock := newRepositoryWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertSettingsQuery)).
		WithArgs(`{"device_id":"x"}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repository.UpsertSettingsDocument(`{"device_id":"x"}`); err != nil {
		t.Fatalf("UpsertSettingsDocument: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryUpsertSettingsDocumentFailure(t *testing.T) {
	repository, mock := newRepositoryWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertSettingsQuery)).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	if err := repository.UpsertSettingsDocument("{}"); err == nil {
		t.Fatal("expected an error")
	}
}
