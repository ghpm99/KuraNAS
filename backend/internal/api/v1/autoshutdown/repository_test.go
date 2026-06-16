package autoshutdown

import (
	"database/sql"
	"regexp"
	"testing"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/autoshutdown"

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
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetSettingsQuery)).
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
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetSettingsQuery)).
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
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertSettingsQuery)).
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

func TestRepositoryGetShutdownTimeMedian(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetShutdownTimeMedianQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"median_seconds", "sample_size"}).AddRow(10800.0, 5))
	mock.ExpectRollback()

	median, sample, err := repo.GetShutdownTimeMedian()
	if err != nil || median != 10800.0 || sample != 5 {
		t.Fatalf("unexpected median result: %v %v %v", median, sample, err)
	}
}

func TestRepositoryGetShutdownTimeMedianNullNoSamples(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetShutdownTimeMedianQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"median_seconds", "sample_size"}).AddRow(nil, 0))
	mock.ExpectRollback()

	median, sample, err := repo.GetShutdownTimeMedian()
	if err != nil || median != 0 || sample != 0 {
		t.Fatalf("expected zero median/sample, got %v %v %v", median, sample, err)
	}
}
