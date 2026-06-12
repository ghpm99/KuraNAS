package email

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/email"

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

var accountColumns = []string{
	"id", "provider", "address", "display_name", "status", "sync_enabled", "last_sync_at", "last_error", "created_at", "updated_at",
}

var accountWithTokenColumns = []string{
	"id", "provider", "address", "display_name", "token_ciphertext", "status", "sync_enabled", "last_sync_at", "last_error", "created_at", "updated_at",
}

func TestRepositoryListAccounts(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListAccountsQuery)).
		WillReturnRows(sqlmock.NewRows(accountColumns).
			AddRow(1, "google", "a@gmail.com", "A", "linked", true, now, "", now, now).
			AddRow(2, "microsoft", "b@hotmail.com", "B", "reauth_required", false, nil, "refresh failed", now, now))
	mock.ExpectRollback()

	accounts, err := repo.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts error: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
	if accounts[0].Provider != ProviderGoogle || accounts[0].LastSyncAt == nil {
		t.Fatalf("unexpected first account: %+v", accounts[0])
	}
	if accounts[1].Status != StatusReauthRequired || accounts[1].LastSyncAt != nil || accounts[1].LastError != "refresh failed" {
		t.Fatalf("unexpected second account: %+v", accounts[1])
	}
	if len(accounts[0].TokenCiphertext) != 0 {
		t.Fatal("listing must not carry token ciphertext")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetAccountByID(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAccountByIDQuery)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows(accountWithTokenColumns).
			AddRow(7, "google", "a@gmail.com", "A", []byte{1, 2, 3}, "linked", true, nil, "", now, now))
	mock.ExpectRollback()

	account, err := repo.GetAccountByID(7)
	if err != nil {
		t.Fatalf("GetAccountByID error: %v", err)
	}
	if account.ID != 7 || string(account.TokenCiphertext) != string([]byte{1, 2, 3}) {
		t.Fatalf("unexpected account: %+v", account)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryUpsertAccountCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertAccountQuery)).
		WithArgs("google", "a@gmail.com", "A", []byte{9, 9}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit()

	id, err := repo.UpsertAccount(AccountModel{
		Provider:        ProviderGoogle,
		Address:         "a@gmail.com",
		DisplayName:     "A",
		TokenCiphertext: []byte{9, 9},
	})
	if err != nil {
		t.Fatalf("UpsertAccount error: %v", err)
	}
	if id != 5 {
		t.Fatalf("expected id 5, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryUpdateAccountTokensCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateAccountTokensQuery)).
		WithArgs(3, []byte{1}, "linked", "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.UpdateAccountTokens(3, []byte{1}, StatusLinked, ""); err != nil {
		t.Fatalf("UpdateAccountTokens error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryUpdateSyncEnabledNotFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateAccountSyncEnabledQuery)).
		WithArgs(99, false).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdateSyncEnabled(99, false)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryDeleteAccountCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAccountQuery)).
		WithArgs(4).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.DeleteAccount(4); err != nil {
		t.Fatalf("DeleteAccount error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
