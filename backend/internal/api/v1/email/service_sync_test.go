package email

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	jobqueries "nas-go/api/pkg/database/queries/jobs"

	"nas-go/api/pkg/crypto"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/mailfetch"

	"github.com/DATA-DOG/go-sqlmock"
)

type fakeFetcher struct {
	messages  []mailfetch.RawMessage
	err       error
	calls     int
	lastSince time.Time
}

func (f *fakeFetcher) ListNewMessages(_ context.Context, _ string, since time.Time, _ int) ([]mailfetch.RawMessage, error) {
	f.calls++
	f.lastSince = since
	if f.err != nil {
		return nil, f.err
	}
	return f.messages, nil
}

func seedAccount(t *testing.T, repo *fakeRepo, cipher *crypto.AESGCM, provider Provider, address string, expiry time.Time) int {
	t.Helper()
	plaintext, err := encodeTokenSet(TokenSet{AccessToken: "at", RefreshToken: "rt", Expiry: expiry})
	if err != nil {
		t.Fatalf("encode tokens: %v", err)
	}
	sealed, err := cipher.Seal(plaintext)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	id, err := repo.UpsertAccount(AccountModel{Provider: provider, Address: address})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	repo.mu.Lock()
	m := repo.accounts[id]
	m.TokenCiphertext = sealed
	m.SyncEnabled = true
	repo.accounts[id] = m
	repo.mu.Unlock()
	return id
}

func TestSyncEnabledAccountsStoresSanitizedAndIsIdempotent(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	service := newTestService(repo, cipher, "http://test.local")
	id := seedAccount(t, repo, cipher, ProviderGoogle, "owner@gmail.com", time.Now().Add(time.Hour))

	fetcher := &fakeFetcher{messages: []mailfetch.RawMessage{{
		ProviderMessageID: "m1",
		SenderAddress:     "alice@example.com",
		Subject:           "Hi",
		Body:              "<p>Hello <script>evil()</script>world</p> see https://Promo.Example.net/x",
		BodyIsHTML:        true,
		ReceivedAt:        time.Now(),
		AuthResults:       mailfetch.AuthResults{DMARC: "pass"},
		Attachments:       []mailfetch.AttachmentMeta{{Filename: "a.pdf", Size: 10}},
	}}}
	service.fetchers = map[Provider]mailfetch.Fetcher{ProviderGoogle: fetcher}

	stats, err := service.SyncEnabledAccounts()
	if err != nil {
		t.Fatalf("SyncEnabledAccounts: %v", err)
	}
	if stats.Accounts != 1 || stats.Fetched != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	if len(repo.messages) != 1 {
		t.Fatalf("expected 1 stored message, got %d", len(repo.messages))
	}
	stored := repo.messages[0]
	if strings.Contains(stored.SanitizedBody, "<") || strings.Contains(stored.SanitizedBody, "evil") {
		t.Fatalf("body not sanitized: %q", stored.SanitizedBody)
	}
	if len(stored.LinkDomains) != 1 || stored.LinkDomains[0] != "promo.example.net" {
		t.Fatalf("unexpected link domains: %v", stored.LinkDomains)
	}
	if stored.Status != MsgStatusPending {
		t.Fatalf("expected pending, got %s", stored.Status)
	}
	if repo.accounts[id].LastSyncAt == nil {
		t.Fatal("expected sync cursor to advance")
	}

	// Second pass: same provider message id must not duplicate.
	stats2, err := service.SyncEnabledAccounts()
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if stats2.Fetched != 0 || len(repo.messages) != 1 {
		t.Fatalf("sync not idempotent: stats=%+v stored=%d", stats2, len(repo.messages))
	}
	if fetcher.lastSince.IsZero() {
		t.Fatal("expected the second pass to pass the stored cursor as since")
	}
}

func TestSyncMarksReauthWithoutAbortingOtherAccounts(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	service := newTestService(repo, cipher, "http://test.local")

	// Expired token + unreachable refresh endpoint => reauth_required.
	seedAccount(t, repo, cipher, ProviderGoogle, "stale@gmail.com", time.Now().Add(-time.Hour))
	// Healthy account on another provider.
	seedAccount(t, repo, cipher, ProviderMicrosoft, "ok@outlook.com", time.Now().Add(time.Hour))

	service.fetchers = map[Provider]mailfetch.Fetcher{
		ProviderGoogle: &fakeFetcher{},
		ProviderMicrosoft: &fakeFetcher{messages: []mailfetch.RawMessage{{
			ProviderMessageID: "x1",
			SenderAddress:     "b@outlook.com",
			ReceivedAt:        time.Now(),
		}}},
	}

	stats, err := service.SyncEnabledAccounts()
	if err != nil {
		t.Fatalf("SyncEnabledAccounts: %v", err)
	}
	if len(stats.ReauthRequired) != 1 || stats.ReauthRequired[0] != "stale@gmail.com" {
		t.Fatalf("expected stale account flagged for reauth, got %+v", stats.ReauthRequired)
	}
	if stats.Fetched != 1 {
		t.Fatalf("healthy account should still sync, got %+v", stats)
	}
}

func TestPrefilterPendingFlagsSpam(t *testing.T) {
	repo := newFakeRepo()
	service := newTestService(repo, testCipher(t), "http://test.local")

	repo.messages = []MessageModel{
		{ID: 1, Status: MsgStatusPending, Subject: "weekly digest", AuthResults: AuthResults{DMARC: "pass"}},
		{ID: 2, Status: MsgStatusPending, Subject: "spoof", AuthResults: AuthResults{DMARC: "fail"}},
	}
	repo.nextMsgID = 3

	flagged, err := service.PrefilterPending()
	if err != nil {
		t.Fatalf("PrefilterPending: %v", err)
	}
	if flagged != 1 {
		t.Fatalf("expected 1 flagged, got %d", flagged)
	}
	if repo.messages[1].Status != MsgStatusPrefilteredSpam || len(repo.messages[1].PrefilterRules) == 0 {
		t.Fatalf("spam message not flagged with rules: %+v", repo.messages[1])
	}
	if repo.messages[0].Status != MsgStatusPending {
		t.Fatalf("clean message should stay pending")
	}
}

func TestPurgeExpired(t *testing.T) {
	repo := newFakeRepo()
	service := NewServiceWithEndpoints(repo, testCipher(t), Config{RetentionDays: 30}, DefaultEndpoints())

	repo.messages = []MessageModel{
		{ID: 1, ReceivedAt: time.Now().Add(-40 * 24 * time.Hour)},
		{ID: 2, ReceivedAt: time.Now()},
	}

	purged, err := service.PurgeExpired()
	if err != nil {
		t.Fatalf("PurgeExpired: %v", err)
	}
	if purged != 1 || len(repo.messages) != 1 || repo.messages[0].ID != 2 {
		t.Fatalf("unexpected purge: removed=%d remaining=%+v", purged, repo.messages)
	}
}

func TestListMessagesClampsPageSizeAndDropsBody(t *testing.T) {
	repo := newFakeRepo()
	service := newTestService(repo, testCipher(t), "http://test.local")
	repo.messages = []MessageModel{{ID: 1, SanitizedBody: "secret", Subject: "Hi", Status: MsgStatusPending}}
	repo.nextMsgID = 2

	page, err := service.ListMessages(0, 1000)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if page.Pagination.Page != 1 || page.Pagination.PageSize != 50 {
		t.Fatalf("page size/number not clamped: %+v", page.Pagination)
	}
	if len(page.Items) != 1 || page.Items[0].Subject != "Hi" {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
}

func TestEnqueueSyncCreatesThreeStepJob(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	service := newTestService(repo, cipher, "http://test.local")
	id := seedAccount(t, repo, cipher, ProviderGoogle, "a@gmail.com", time.Now().Add(time.Hour))

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	service.SetJobsDispatcher(jobs.NewRepository(database.NewDbContext(db)))

	jobCols := []string{"id", "type", "priority", "scope", "status", "created_at", "started_at", "ended_at", "cancel_requested", "last_error"}
	stepCols := []string{"id", "job_id", "type", "status", "depends_on", "attempts", "max_attempts", "last_error", "progress", "payload", "created_at", "started_at", "ended_at"}
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(jobqueries.InsertJobQuery)).
		WillReturnRows(sqlmock.NewRows(jobCols).
			AddRow(77, "email_sync", "normal", nil, "queued", now, nil, nil, false, ""))
	mock.ExpectQuery(regexp.QuoteMeta(jobqueries.InsertStepQuery)).
		WillReturnRows(sqlmock.NewRows(stepCols).
			AddRow(1, 77, "email_fetch", "queued", []byte("[]"), 0, 1, "", 0, nil, now, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta(jobqueries.InsertStepQuery)).
		WillReturnRows(sqlmock.NewRows(stepCols).
			AddRow(2, 77, "email_prefilter", "queued", []byte("[1]"), 0, 1, "", 0, nil, now, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta(jobqueries.InsertStepQuery)).
		WillReturnRows(sqlmock.NewRows(stepCols).
			AddRow(3, 77, "email_analyze", "queued", []byte("[2]"), 0, 1, "", 0, nil, now, nil, nil))
	mock.ExpectCommit()

	jobID, err := service.EnqueueSync(id)
	if err != nil {
		t.Fatalf("EnqueueSync: %v", err)
	}
	if jobID != 77 {
		t.Fatalf("expected job id 77, got %d", jobID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEnqueueSyncWithoutDispatcher(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	service := newTestService(repo, cipher, "http://test.local")
	id := seedAccount(t, repo, cipher, ProviderGoogle, "a@gmail.com", time.Now().Add(time.Hour))

	if _, err := service.EnqueueSync(id); !errors.Is(err, ErrSyncUnavailable) {
		t.Fatalf("expected ErrSyncUnavailable, got %v", err)
	}
	if _, err := service.EnqueueSync(999); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound for unknown account, got %v", err)
	}
}
