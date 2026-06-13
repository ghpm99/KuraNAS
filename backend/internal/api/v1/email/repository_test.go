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

func TestRepositoryUpdateAccountLastSyncCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateAccountLastSyncQuery)).
		WithArgs(3, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.UpdateAccountLastSync(3, now); err != nil {
		t.Fatalf("UpdateAccountLastSync error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryInsertMessage(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertMessageQuery)).
		WithArgs(1, "m1", "", "", "", "", "", now,
			[]byte(`{"spf":"","dkim":"","dmarc":""}`), []byte("[]"), []byte("[]"), []byte("[]"), "pending").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	inserted, err := repo.InsertMessage(MessageModel{AccountID: 1, ProviderMessageID: "m1", ReceivedAt: now})
	if err != nil || !inserted {
		t.Fatalf("InsertMessage: inserted=%v err=%v", inserted, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryInsertMessageConflictIsNotInserted(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertMessageQuery)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectCommit()

	inserted, err := repo.InsertMessage(MessageModel{AccountID: 1, ProviderMessageID: "dup", ReceivedAt: now})
	if err != nil {
		t.Fatalf("conflict must not error: %v", err)
	}
	if inserted {
		t.Fatal("conflicting message must report inserted=false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

var messageColumns = []string{
	"id", "account_id", "sender_name", "sender_address", "subject", "snippet", "received_at", "status", "created_at",
	"verdict", "importance", "summary",
}

func TestRepositoryListMessages(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListMessagesQuery)).
		WithArgs(3, 0).
		WillReturnRows(sqlmock.NewRows(messageColumns).
			AddRow(1, 5, "Alice", "a@example.com", "Hi", "snippet", now, "pending", now, "", "", "").
			AddRow(2, 5, "Bob", "b@example.com", "Yo", "snip2", now, "analyzed", now, "legitimate", "high", "A short summary."))
	mock.ExpectRollback()

	page, err := repo.ListMessages(1, 2)
	if err != nil {
		t.Fatalf("ListMessages error: %v", err)
	}
	if len(page.Items) != 2 || page.Items[0].SenderAddress != "a@example.com" {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryListPendingMessages(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListPendingMessagesQuery)).
		WithArgs(500).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sender_address", "subject", "auth_results", "attachment_meta", "link_domains"}).
			AddRow(1, "a@example.com", "Hi",
				[]byte(`{"spf":"pass","dkim":"pass","dmarc":"fail"}`),
				[]byte(`[{"filename":"x.pdf","mime":"application/pdf","size":10}]`),
				[]byte(`["promo.net"]`)))
	mock.ExpectRollback()

	pending, err := repo.ListPendingMessages(500)
	if err != nil {
		t.Fatalf("ListPendingMessages error: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
	m := pending[0]
	if m.AuthResults.DMARC != "fail" || len(m.Attachments) != 1 || m.Attachments[0].Filename != "x.pdf" {
		t.Fatalf("unexpected decoded message: %+v", m)
	}
	if len(m.LinkDomains) != 1 || m.LinkDomains[0] != "promo.net" {
		t.Fatalf("unexpected link domains: %v", m.LinkDomains)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryUpdateMessagePrefilter(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateMessagePrefilterQuery)).
		WithArgs(8, "prefiltered_spam", []byte(`["dmarc_fail"]`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.UpdateMessagePrefilter(8, MsgStatusPrefilteredSpam, []string{"dmarc_fail"}); err != nil {
		t.Fatalf("UpdateMessagePrefilter error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryPurgeMessagesBefore(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	cutoff := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.PurgeMessagesBeforeQuery)).
		WithArgs(cutoff).
		WillReturnResult(sqlmock.NewResult(0, 4))
	mock.ExpectCommit()

	removed, err := repo.PurgeMessagesBefore(cutoff)
	if err != nil {
		t.Fatalf("PurgeMessagesBefore error: %v", err)
	}
	if removed != 4 {
		t.Fatalf("expected 4 removed, got %d", removed)
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
