package email

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"nas-go/api/pkg/crypto"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

// fakeRepo is an in-memory RepositoryInterface for service tests.
type fakeRepo struct {
	mu       sync.Mutex
	accounts map[int]AccountModel
	nextID   int

	messages   []MessageModel
	nextMsgID  int
	lastSyncAt map[int]time.Time
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{accounts: map[int]AccountModel{}, nextID: 1, nextMsgID: 1, lastSyncAt: map[int]time.Time{}}
}

func (f *fakeRepo) GetDbContext() *database.DbContext { return nil }

func (f *fakeRepo) ListAccounts() ([]AccountModel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var models []AccountModel
	for _, m := range f.accounts {
		listed := m
		listed.TokenCiphertext = nil
		models = append(models, listed)
	}
	return models, nil
}

func (f *fakeRepo) GetAccountByID(id int) (AccountModel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.accounts[id]
	if !ok {
		return AccountModel{}, sql.ErrNoRows
	}
	return m, nil
}

func (f *fakeRepo) UpsertAccount(model AccountModel) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.accounts {
		if existing.Provider == model.Provider && existing.Address == model.Address {
			model.ID = id
			model.Status = StatusLinked
			f.accounts[id] = model
			return id, nil
		}
	}
	model.ID = f.nextID
	model.Status = StatusLinked
	f.accounts[model.ID] = model
	f.nextID++
	return model.ID, nil
}

func (f *fakeRepo) UpdateAccountTokens(id int, tokenCiphertext []byte, status AccountStatus, lastError string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.accounts[id]
	if !ok {
		return sql.ErrNoRows
	}
	m.TokenCiphertext = tokenCiphertext
	m.Status = status
	m.LastError = lastError
	f.accounts[id] = m
	return nil
}

func (f *fakeRepo) UpdateSyncEnabled(id int, enabled bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.accounts[id]
	if !ok {
		return sql.ErrNoRows
	}
	m.SyncEnabled = enabled
	f.accounts[id] = m
	return nil
}

func (f *fakeRepo) DeleteAccount(id int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.accounts[id]; !ok {
		return sql.ErrNoRows
	}
	delete(f.accounts, id)
	return nil
}

func (f *fakeRepo) UpdateAccountLastSync(id int, syncedAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.accounts[id]
	if !ok {
		return sql.ErrNoRows
	}
	m.Status = StatusLinked
	m.LastError = ""
	m.LastSyncAt = &syncedAt
	f.accounts[id] = m
	f.lastSyncAt[id] = syncedAt
	return nil
}

func (f *fakeRepo) InsertMessage(message MessageModel) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.messages {
		if existing.AccountID == message.AccountID && existing.ProviderMessageID == message.ProviderMessageID {
			return false, nil
		}
	}
	message.ID = f.nextMsgID
	f.nextMsgID++
	if message.Status == "" {
		message.Status = MsgStatusPending
	}
	f.messages = append(f.messages, message)
	return true, nil
}

func (f *fakeRepo) ListMessages(page, pageSize int) (utils.PaginationResponse[MessageModel], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	response := utils.PaginationResponse[MessageModel]{
		Items:      append([]MessageModel{}, f.messages...),
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}
	response.UpdatePagination()
	return response, nil
}

func (f *fakeRepo) ListPendingMessages(limit int) ([]MessageModel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var pending []MessageModel
	for _, m := range f.messages {
		if m.Status == MsgStatusPending {
			pending = append(pending, m)
		}
		if len(pending) >= limit {
			break
		}
	}
	return pending, nil
}

func (f *fakeRepo) UpdateMessagePrefilter(id int, status MessageStatus, rules []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.messages {
		if f.messages[i].ID == id {
			f.messages[i].Status = status
			f.messages[i].PrefilterRules = rules
			return nil
		}
	}
	return sql.ErrNoRows
}

func (f *fakeRepo) PurgeMessagesBefore(cutoff time.Time) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	kept := f.messages[:0:0]
	removed := 0
	for _, m := range f.messages {
		if m.ReceivedAt.Before(cutoff) {
			removed++
			continue
		}
		kept = append(kept, m)
	}
	f.messages = kept
	return removed, nil
}

func (f *fakeRepo) account(t *testing.T, id int) AccountModel {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.accounts[id]
	if !ok {
		t.Fatalf("account %d not found in fake repo", id)
	}
	return m
}

func testCipher(t *testing.T) *crypto.AESGCM {
	t.Helper()
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	c, err := crypto.NewAESGCM(key)
	if err != nil {
		t.Fatalf("NewAESGCM: %v", err)
	}
	return c
}

func fakeIDToken(t *testing.T, claims map[string]string) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encode := base64.RawURLEncoding.EncodeToString
	return encode([]byte(`{"alg":"none"}`)) + "." + encode(payload) + ".sig"
}

func newTestService(repo RepositoryInterface, cipher *crypto.AESGCM, serverURL string) *Service {
	return NewServiceWithEndpoints(repo, cipher, Config{
		GoogleClientID:     "gid",
		GoogleClientSecret: "gsecret",
		MicrosoftClientID:  "msid",
	}, Endpoints{
		GoogleAuth:          serverURL + "/google/auth",
		GoogleToken:         serverURL + "/google/token",
		MicrosoftDeviceCode: serverURL + "/ms/devicecode",
		MicrosoftToken:      serverURL + "/ms/token",
	})
}

func TestGoogleAuthURLCarriesExactScopesAndPKCE(t *testing.T) {
	service := newTestService(newFakeRepo(), testCipher(t), "http://test.local")

	dto, err := service.GoogleAuthURL()
	if err != nil {
		t.Fatalf("GoogleAuthURL: %v", err)
	}

	parsed, err := url.Parse(dto.AuthURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	q := parsed.Query()

	if got := q.Get("scope"); got != "https://www.googleapis.com/auth/gmail.readonly openid email" {
		t.Fatalf("unexpected scope: %q", got)
	}
	if strings.Contains(q.Get("scope"), "send") || strings.Contains(q.Get("scope"), "modify") {
		t.Fatal("scope must be read-only")
	}
	if q.Get("code_challenge") == "" || q.Get("code_challenge_method") != "S256" {
		t.Fatal("missing PKCE challenge")
	}
	if q.Get("state") == "" || q.Get("access_type") != "offline" {
		t.Fatal("missing state or offline access")
	}
}

func TestGoogleAuthURLRequiresClientConfig(t *testing.T) {
	service := NewServiceWithEndpoints(newFakeRepo(), testCipher(t), Config{}, DefaultEndpoints())
	if _, err := service.GoogleAuthURL(); !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("expected ErrProviderNotConfigured, got %v", err)
	}
}

func TestHandleGoogleCallbackPersistsSealedTokens(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	idToken := fakeIDToken(t, map[string]string{"email": "Owner@Gmail.com", "name": "Owner"})

	var gotVerifier string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/google/token" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = r.ParseForm()
		gotVerifier = r.Form.Get("code_verifier")
		if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "the-code" {
			t.Fatalf("unexpected form: %v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at-1",
			"refresh_token": "rt-1",
			"expires_in":    3600,
			"id_token":      idToken,
		})
	}))
	defer server.Close()

	service := newTestService(repo, cipher, server.URL)

	dto, err := service.GoogleAuthURL()
	if err != nil {
		t.Fatalf("GoogleAuthURL: %v", err)
	}
	state := mustQueryParam(t, dto.AuthURL, "state")

	if err := service.HandleGoogleCallback(state, "the-code"); err != nil {
		t.Fatalf("HandleGoogleCallback: %v", err)
	}
	if gotVerifier == "" {
		t.Fatal("PKCE verifier was not sent on exchange")
	}

	account := repo.account(t, 1)
	if account.Provider != ProviderGoogle || account.Address != "owner@gmail.com" {
		t.Fatalf("unexpected account: %+v", account)
	}
	if bytes.Contains(account.TokenCiphertext, []byte("at-1")) || bytes.Contains(account.TokenCiphertext, []byte("rt-1")) {
		t.Fatal("tokens stored in plaintext")
	}

	plaintext, err := cipher.Open(account.TokenCiphertext)
	if err != nil {
		t.Fatalf("stored blob does not decrypt: %v", err)
	}
	var tokens TokenSet
	if err := json.Unmarshal(plaintext, &tokens); err != nil {
		t.Fatalf("decode tokens: %v", err)
	}
	if tokens.AccessToken != "at-1" || tokens.RefreshToken != "rt-1" {
		t.Fatalf("unexpected tokens: %+v", tokens)
	}
}

func TestHandleGoogleCallbackRejectsUnknownState(t *testing.T) {
	service := newTestService(newFakeRepo(), testCipher(t), "http://test.local")
	if err := service.HandleGoogleCallback("nope", "code"); !errors.Is(err, ErrInvalidOAuthState) {
		t.Fatalf("expected ErrInvalidOAuthState, got %v", err)
	}
}

func TestMicrosoftDeviceCodeFlowLinksAccount(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	idToken := fakeIDToken(t, map[string]string{"preferred_username": "owner@hotmail.com", "name": "Owner"})

	var scopeAsked string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		switch r.URL.Path {
		case "/ms/devicecode":
			scopeAsked = r.Form.Get("scope")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      "dev-1",
				"user_code":        "ABC123",
				"verification_uri": "https://microsoft.com/devicelogin",
				"expires_in":       900,
				"interval":         1,
			})
		case "/ms/token":
			if r.Form.Get("device_code") != "dev-1" {
				t.Fatalf("unexpected device code: %v", r.Form)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "ms-at",
				"refresh_token": "ms-rt",
				"expires_in":    3600,
				"id_token":      idToken,
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := newTestService(repo, cipher, server.URL)

	dto, err := service.StartMicrosoftDeviceCode()
	if err != nil {
		t.Fatalf("StartMicrosoftDeviceCode: %v", err)
	}
	if dto.UserCode != "ABC123" || dto.VerificationURI == "" {
		t.Fatalf("unexpected device code dto: %+v", dto)
	}
	if !strings.Contains(scopeAsked, "Mail.Read") || !strings.Contains(scopeAsked, "offline_access") {
		t.Fatalf("unexpected scope: %q", scopeAsked)
	}
	if strings.Contains(scopeAsked, "Mail.Send") || strings.Contains(scopeAsked, "ReadWrite") {
		t.Fatal("scope must be read-only")
	}

	waitForStatus(t, service, DeviceCodeLinked)

	account := repo.account(t, 1)
	if account.Provider != ProviderMicrosoft || account.Address != "owner@hotmail.com" {
		t.Fatalf("unexpected account: %+v", account)
	}
	if bytes.Contains(account.TokenCiphertext, []byte("ms-at")) {
		t.Fatal("tokens stored in plaintext")
	}
}

func TestMicrosoftDeviceCodePendingThenExpired(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ms/devicecode":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code": "dev-2", "user_code": "X", "verification_uri": "v", "expires_in": 900, "interval": 1,
			})
		case "/ms/token":
			calls++
			w.WriteHeader(http.StatusBadRequest)
			if calls == 1 {
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "expired_token"})
		}
	}))
	defer server.Close()

	service := newTestService(newFakeRepo(), testCipher(t), server.URL)
	if _, err := service.StartMicrosoftDeviceCode(); err != nil {
		t.Fatalf("StartMicrosoftDeviceCode: %v", err)
	}

	waitForStatus(t, service, DeviceCodeExpired)
	if calls < 2 {
		t.Fatalf("expected at least 2 polls, got %d", calls)
	}
}

func TestValidAccessTokenReturnsUnexpiredWithoutHTTP(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	service := newTestService(repo, cipher, "http://test.local")

	sealed := mustSeal(t, cipher, TokenSet{AccessToken: "fresh", RefreshToken: "r", Expiry: time.Now().Add(time.Hour)})
	id, _ := repo.UpsertAccount(AccountModel{Provider: ProviderGoogle, Address: "a@gmail.com", TokenCiphertext: sealed})

	token, err := service.ValidAccessToken(id)
	if err != nil {
		t.Fatalf("ValidAccessToken: %v", err)
	}
	if token != "fresh" {
		t.Fatalf("expected fresh token, got %q", token)
	}
}

func TestValidAccessTokenRefreshesExpired(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.URL.Path != "/google/token" || r.Form.Get("grant_type") != "refresh_token" {
			t.Fatalf("unexpected request: %s %v", r.URL.Path, r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "renewed", "expires_in": 3600})
	}))
	defer server.Close()

	service := newTestService(repo, cipher, server.URL)

	sealed := mustSeal(t, cipher, TokenSet{AccessToken: "old", RefreshToken: "keep-me", Expiry: time.Now().Add(-time.Hour)})
	id, _ := repo.UpsertAccount(AccountModel{Provider: ProviderGoogle, Address: "a@gmail.com", TokenCiphertext: sealed})

	token, err := service.ValidAccessToken(id)
	if err != nil {
		t.Fatalf("ValidAccessToken: %v", err)
	}
	if token != "renewed" {
		t.Fatalf("expected renewed token, got %q", token)
	}

	account := repo.account(t, id)
	if account.Status != StatusLinked {
		t.Fatalf("expected linked status, got %s", account.Status)
	}
	plaintext, _ := cipher.Open(account.TokenCiphertext)
	var tokens TokenSet
	_ = json.Unmarshal(plaintext, &tokens)
	if tokens.AccessToken != "renewed" || tokens.RefreshToken != "keep-me" {
		t.Fatalf("refresh token must be kept when the response omits it: %+v", tokens)
	}
}

func TestValidAccessTokenMarksReauthRequired(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
	}))
	defer server.Close()

	service := newTestService(repo, cipher, server.URL)

	sealed := mustSeal(t, cipher, TokenSet{AccessToken: "old", RefreshToken: "dead", Expiry: time.Now().Add(-time.Hour)})
	id, _ := repo.UpsertAccount(AccountModel{Provider: ProviderMicrosoft, Address: "a@hotmail.com", TokenCiphertext: sealed})

	if _, err := service.ValidAccessToken(id); !errors.Is(err, ErrReauthRequired) {
		t.Fatalf("expected ErrReauthRequired, got %v", err)
	}

	account := repo.account(t, id)
	if account.Status != StatusReauthRequired {
		t.Fatalf("expected reauth_required, got %s", account.Status)
	}
	if !strings.Contains(account.LastError, "invalid_grant") {
		t.Fatalf("expected oauth error code in last_error, got %q", account.LastError)
	}
	if strings.Contains(account.LastError, "dead") {
		t.Fatal("last_error must not contain tokens")
	}
}

func TestValidAccessTokenUnknownAccount(t *testing.T) {
	service := newTestService(newFakeRepo(), testCipher(t), "http://test.local")
	if _, err := service.ValidAccessToken(42); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestPostFormRejectsHostOutsideAllowlist(t *testing.T) {
	service := newTestService(newFakeRepo(), testCipher(t), "http://allowed.local")
	if _, _, err := service.postFormRaw("http://evil.example.com/token", url.Values{}); !errors.Is(err, ErrHostNotAllowed) {
		t.Fatalf("expected ErrHostNotAllowed, got %v", err)
	}
}

func TestDefaultAllowlistIsTheFixedProviderSet(t *testing.T) {
	allowed := allowedHostsFor(DefaultEndpoints())
	for _, host := range []string{"accounts.google.com", "oauth2.googleapis.com", "login.microsoftonline.com"} {
		if !allowed[host] {
			t.Fatalf("expected %s in allowlist", host)
		}
	}
	if len(allowed) != 3 {
		t.Fatalf("allowlist must stay minimal, got %v", allowed)
	}
}

func TestCrudServiceMethods(t *testing.T) {
	repo := newFakeRepo()
	cipher := testCipher(t)
	service := newTestService(repo, cipher, "http://test.local")

	id, _ := repo.UpsertAccount(AccountModel{Provider: ProviderGoogle, Address: "a@gmail.com", TokenCiphertext: []byte{1}})

	dtos, err := service.ListAccounts()
	if err != nil || len(dtos) != 1 {
		t.Fatalf("ListAccounts: %v / %d", err, len(dtos))
	}

	if err := service.SetSyncEnabled(id, false); err != nil {
		t.Fatalf("SetSyncEnabled: %v", err)
	}
	if repo.account(t, id).SyncEnabled {
		t.Fatal("sync should be disabled")
	}
	if err := service.SetSyncEnabled(99, true); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}

	if err := service.DeleteAccount(id); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	if err := service.DeleteAccount(id); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func mustSeal(t *testing.T, cipher *crypto.AESGCM, tokens TokenSet) []byte {
	t.Helper()
	plaintext, err := json.Marshal(tokens)
	if err != nil {
		t.Fatalf("marshal tokens: %v", err)
	}
	sealed, err := cipher.Seal(plaintext)
	if err != nil {
		t.Fatalf("seal tokens: %v", err)
	}
	return sealed
}

func mustQueryParam(t *testing.T, rawURL string, key string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	value := parsed.Query().Get(key)
	if value == "" {
		t.Fatalf("missing %s in %s", key, rawURL)
	}
	return value
}

func waitForStatus(t *testing.T, service *Service, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if service.MicrosoftDeviceCodeStatus().Status == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("device link never reached %q (last: %q)", want, service.MicrosoftDeviceCodeStatus().Status)
}
