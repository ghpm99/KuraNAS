package configuration

import (
	"database/sql"
	"regexp"
	"testing"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/configuration"

	"github.com/DATA-DOG/go-sqlmock"
)

func newConfigurationRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestConfigurationRepositoryGetSettingsDocument(t *testing.T) {
	repo, mock, db := newConfigurationRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetSettingQuery)).
		WithArgs(settingsStorageKey).
		WillReturnRows(sqlmock.NewRows([]string{"setting_key", "setting_value"}).AddRow(settingsStorageKey, `{"language":{"current":"en-US"}}`))
	mock.ExpectRollback()

	payload, err := repo.GetSettingsDocument(settingsStorageKey)
	if err != nil {
		t.Fatalf("GetSettingsDocument returned error: %v", err)
	}
	if payload == "" {
		t.Fatalf("expected payload")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestConfigurationRepositoryUpsertSettingsDocument(t *testing.T) {
	repo, mock, db := newConfigurationRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertSettingQuery)).
		WithArgs(settingsStorageKey, `{}`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.UpsertSettingsDocument(tx, settingsStorageKey, `{}`)
	})
	if err != nil {
		t.Fatalf("UpsertSettingsDocument returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
