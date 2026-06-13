package mailfetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGuardedClientGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("expected bearer token, got %q", got)
		}
		w.Write([]byte(`{"value":"ok"}`))
	}))
	defer server.Close()

	host := mustHost(t, server.URL)
	client := NewGuardedClient(host)

	var out struct {
		Value string `json:"value"`
	}
	if err := client.GetJSON(context.Background(), server.URL, "tok", &out); err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if out.Value != "ok" {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestGuardedClientRejectsForeignHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("foreign host must never be contacted")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Allow some unrelated host, then try to reach the (different) test server.
	client := NewGuardedClient("api.example.invalid")

	var out any
	err := client.GetJSON(context.Background(), server.URL, "tok", &out)
	if !errors.Is(err, ErrHostNotAllowed) {
		t.Fatalf("expected ErrHostNotAllowed, got %v", err)
	}
}

func TestGuardedClientNon2xxIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewGuardedClient(mustHost(t, server.URL))
	var out any
	if err := client.GetJSON(context.Background(), server.URL, "tok", &out); err == nil {
		t.Fatal("expected error on 403")
	}
}

func mustHost(t *testing.T, raw string) string {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return parsed.Hostname()
}
