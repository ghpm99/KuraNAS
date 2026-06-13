package mailfetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// maxResponseBytes caps how much of any provider response is read into memory.
// Bodies are truncated to 16 KB during sanitization anyway; this is the
// transport-level ceiling against a hostile or runaway response.
const maxResponseBytes = 4 << 20 // 4 MiB

// GuardedClient issues authenticated GET requests to a single provider,
// refusing any URL — including a redirect hop — whose host is not in a fixed
// allowlist, before a single byte leaves the process. This is the enforcement
// point for the "talk only to the provider host" hard rule.
type GuardedClient struct {
	httpClient *http.Client
	allowed    map[string]bool
}

// NewGuardedClient builds a client allowing exactly the given hosts. Redirects
// are re-checked against the same allowlist, so a 3xx pointing elsewhere is
// rejected rather than followed.
func NewGuardedClient(allowedHosts ...string) *GuardedClient {
	allowed := make(map[string]bool, len(allowedHosts))
	for _, host := range allowedHosts {
		if host != "" {
			allowed[host] = true
		}
	}

	client := &GuardedClient{allowed: allowed}
	client.httpClient = &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if !client.allowed[req.URL.Hostname()] {
				return ErrHostNotAllowed
			}
			return nil
		},
	}
	return client
}

// GetJSON fetches rawURL with a bearer token and decodes the JSON body into out.
// The host is verified against the allowlist before the request is made.
func (g *GuardedClient) GetJSON(ctx context.Context, rawURL, accessToken string, out any) error {
	body, err := g.get(ctx, rawURL, accessToken)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("mailfetch: decode response: %w", err)
	}
	return nil
}

func (g *GuardedClient) get(ctx context.Context, rawURL, accessToken string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || !g.allowed[parsed.Hostname()] {
		return nil, ErrHostNotAllowed
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("mailfetch: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mailfetch: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("mailfetch: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mailfetch: provider returned status %d", resp.StatusCode)
	}
	return body, nil
}
