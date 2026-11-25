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

func TestRepositoryCloudSettingsAndRoomNamesDocuments(t *testing.T) {
	repository, mock := newRepositoryWithMock(t)
	documents := []struct {
		name   string
		get    func() (string, bool, error)
		upsert func(string) error
		getSQL string
		putSQL string
	}{
		{"cloud settings", repository.GetCloudSettingsDocument, repository.UpsertCloudSettingsDocument, queries.GetCloudSettingsQuery, queries.UpsertCloudSettingsQuery},
		{"room names", repository.GetRoomNamesDocument, repository.UpsertRoomNamesDocument, queries.GetRoomNamesQuery, queries.UpsertRoomNamesQuery},
	}
	for _, document := range documents {
		t.Run(document.name, func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(document.getSQL)).
				WillReturnRows(sqlmock.NewRows([]string{"setting_value"}).AddRow(`{"a":1}`))
			mock.ExpectRollback()
			stored, isFound, err := document.get()
			if err != nil || !isFound || stored != `{"a":1}` {
				t.Fatalf("unexpected read: %q %v %v", stored, isFound, err)
			}

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(document.getSQL)).WillReturnError(sql.ErrNoRows)
			mock.ExpectRollback()
			if _, isFound, err := document.get(); err != nil || isFound {
				t.Fatalf("expected a clean miss, got %v %v", isFound, err)
			}

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(document.getSQL)).WillReturnError(errors.New("db down"))
			mock.ExpectRollback()
			if _, _, err := document.get(); err == nil {
				t.Fatal("expected a read failure")
			}

			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(document.putSQL)).WithArgs(`{"a":2}`).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			if err := document.upsert(`{"a":2}`); err != nil {
				t.Fatal(err)
			}

			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(document.putSQL)).WillReturnError(errors.New("db down"))
			mock.ExpectRollback()
			if err := document.upsert("{}"); err == nil {
				t.Fatal("expected a write failure")
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
