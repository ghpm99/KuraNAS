package email

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"nas-go/api/pkg/crypto"
)

var (
	ErrAccountNotFound       = errors.New("email: account not found")
	ErrProviderNotConfigured = errors.New("email: oauth client not configured")
	ErrInvalidOAuthState     = errors.New("email: unknown or expired oauth state")
	ErrReauthRequired        = errors.New("email: refresh token rejected, reauthorization required")
	ErrHostNotAllowed        = errors.New("email: host not in the oauth allowlist")
	ErrNoDeviceLink          = errors.New("email: no device-code link in progress")
)

// Config carries the OAuth client settings (from env, set in the composition
// root). Secrets never leave this struct.
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	MicrosoftClientID  string
}

// Endpoints groups every external URL the service may call. Production always
// uses DefaultEndpoints; tests point them at an httptest server.
type Endpoints struct {
	GoogleAuth          string
	GoogleToken         string
	MicrosoftDeviceCode string
	MicrosoftToken      string
}

func DefaultEndpoints() Endpoints {
	return Endpoints{
		GoogleAuth:          "https://accounts.google.com/o/oauth2/v2/auth",
		GoogleToken:         "https://oauth2.googleapis.com/token",
		MicrosoftDeviceCode: "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode",
		MicrosoftToken:      "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
	}
}

// allowedHostsFor is the fixed allowlist of hosts the e-mail OAuth client may
// reach (hard rule of the e-mail feature). Graph/Gmail API hosts only join in
// task 15.
func allowedHostsFor(endpoints Endpoints) map[string]bool {
	hosts := make(map[string]bool)
	for _, raw := range []string{endpoints.GoogleAuth, endpoints.GoogleToken, endpoints.MicrosoftDeviceCode, endpoints.MicrosoftToken} {
		if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
			hosts[u.Hostname()] = true
		}
	}
	return hosts
}

type pkceState struct {
	verifier  string
	expiresAt time.Time
}

type deviceLinkState struct {
	status string
}

type Service struct {
	repository RepositoryInterface
	cipher     *crypto.AESGCM
	config     Config
	endpoints  Endpoints
	allowed    map[string]bool
	httpClient *http.Client

	mu         sync.Mutex
	pkceStates map[string]pkceState
	deviceLink deviceLinkState
}

func NewService(repository RepositoryInterface, cipher *crypto.AESGCM, config Config) *Service {
	return NewServiceWithEndpoints(repository, cipher, config, DefaultEndpoints())
}

// NewServiceWithEndpoints exists for tests; production wiring uses NewService.
func NewServiceWithEndpoints(repository RepositoryInterface, cipher *crypto.AESGCM, config Config, endpoints Endpoints) *Service {
	return &Service{
		repository: repository,
		cipher:     cipher,
		config:     config,
		endpoints:  endpoints,
		allowed:    allowedHostsFor(endpoints),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		pkceStates: make(map[string]pkceState),
		deviceLink: deviceLinkState{status: DeviceCodeIdle},
	}
}

func (s *Service) ListAccounts() ([]AccountDto, error) {
	models, err := s.repository.ListAccounts()
	if err != nil {
		return nil, err
	}

	dtos := make([]AccountDto, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, model.toDto())
	}
	return dtos, nil
}

func (s *Service) DeleteAccount(id int) error {
	err := s.repository.DeleteAccount(id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	return err
}

func (s *Service) SetSyncEnabled(id int, enabled bool) error {
	err := s.repository.UpdateSyncEnabled(id, enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	return err
}

// ValidAccessToken returns a usable access token for the account, refreshing
// and re-sealing the token set when it is expired (or about to expire). A
// rejected refresh marks the account reauth_required.
func (s *Service) ValidAccessToken(accountID int) (string, error) {
	account, err := s.repository.GetAccountByID(accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrAccountNotFound
	}
	if err != nil {
		return "", err
	}

	tokens, err := s.openTokens(account.TokenCiphertext)
	if err != nil {
		return "", err
	}

	if time.Until(tokens.Expiry) > time.Minute {
		return tokens.AccessToken, nil
	}

	refreshed, refreshErr := s.refreshTokens(account.Provider, tokens)
	if refreshErr != nil {
		// The OAuth error code is safe to persist; tokens never are.
		_ = s.repository.UpdateAccountTokens(account.ID, account.TokenCiphertext, StatusReauthRequired, refreshErr.Error())
		return "", ErrReauthRequired
	}

	sealed, err := s.sealTokens(refreshed)
	if err != nil {
		return "", err
	}
	if err := s.repository.UpdateAccountTokens(account.ID, sealed, StatusLinked, ""); err != nil {
		return "", err
	}

	return refreshed.AccessToken, nil
}

func (s *Service) refreshTokens(provider Provider, tokens TokenSet) (TokenSet, error) {
	switch provider {
	case ProviderGoogle:
		return s.refreshGoogleTokens(tokens)
	case ProviderMicrosoft:
		return s.refreshMicrosoftTokens(tokens)
	default:
		return TokenSet{}, fmt.Errorf("email: unknown provider %q", provider)
	}
}

func (s *Service) sealTokens(tokens TokenSet) ([]byte, error) {
	plaintext, err := encodeTokenSet(tokens)
	if err != nil {
		return nil, err
	}
	return s.cipher.Seal(plaintext)
}

func (s *Service) openTokens(sealed []byte) (TokenSet, error) {
	plaintext, err := s.cipher.Open(sealed)
	if err != nil {
		return TokenSet{}, err
	}
	return decodeTokenSet(plaintext)
}

// persistLinkedAccount seals the token set and upserts the account row;
// re-linking an existing (provider, address) pair just rotates its tokens.
func (s *Service) persistLinkedAccount(provider Provider, address, displayName string, tokens TokenSet) error {
	sealed, err := s.sealTokens(tokens)
	if err != nil {
		return err
	}

	_, err = s.repository.UpsertAccount(AccountModel{
		Provider:        provider,
		Address:         strings.ToLower(strings.TrimSpace(address)),
		DisplayName:     displayName,
		TokenCiphertext: sealed,
	})
	return err
}

// tokenResponse is the common OAuth2 token endpoint payload (Google and
// Microsoft both follow RFC 6749 here).
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	IDToken          string `json:"id_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (r tokenResponse) toTokenSet(previousRefresh string) TokenSet {
	refresh := r.RefreshToken
	if refresh == "" {
		refresh = previousRefresh
	}
	return TokenSet{
		AccessToken:  r.AccessToken,
		RefreshToken: refresh,
		Expiry:       time.Now().Add(time.Duration(r.ExpiresIn) * time.Second),
	}
}

// postFormRaw submits a form to an OAuth endpoint, enforcing the fixed host
// allowlist before any byte leaves the process.
func (s *Service) postFormRaw(rawURL string, form url.Values) ([]byte, int, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || !s.allowed[parsed.Hostname()] {
		return nil, 0, ErrHostNotAllowed
	}

	resp, err := s.httpClient.PostForm(rawURL, form)
	if err != nil {
		return nil, 0, fmt.Errorf("email: oauth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("email: oauth response read failed: %w", err)
	}
	return body, resp.StatusCode, nil
}

// postForm is postFormRaw decoded into the RFC 6749 token response shape.
func (s *Service) postForm(rawURL string, form url.Values) (tokenResponse, int, error) {
	body, status, err := s.postFormRaw(rawURL, form)
	if err != nil {
		return tokenResponse{}, status, err
	}

	var parsedBody tokenResponse
	if err := json.Unmarshal(body, &parsedBody); err != nil {
		return tokenResponse{}, status, fmt.Errorf("email: oauth response parse failed: %w", err)
	}
	return parsedBody, status, nil
}

// oauthError wraps a provider error code. It carries only the RFC 6749 error
// code — never tokens or raw response bodies.
func oauthError(stage string, code string) error {
	if code == "" {
		code = "unknown_error"
	}
	return fmt.Errorf("email: %s failed: %s", stage, code)
}

// emailFromIDToken extracts the account address from an OIDC id_token without
// signature verification — the token arrived directly from the provider's
// token endpoint over TLS, so its origin is already authenticated.
func emailFromIDToken(idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return "", errors.New("email: malformed id_token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("email: id_token payload decode failed: %w", err)
	}

	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("email: id_token claims parse failed: %w", err)
	}

	address := claims.Email
	if address == "" {
		address = claims.PreferredUsername
	}
	if address == "" || !strings.Contains(address, "@") {
		return "", errors.New("email: id_token carries no address claim")
	}
	return address, nil
}

func displayNameFromIDToken(idToken string) string {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.Name
}

// randomToken returns a URL-safe random string for states and PKCE verifiers.
func randomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
