package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"nas-go/api/pkg/ai"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestProvider(serverURL, apiKey, model string) *Provider {
	return &Provider{
		apiKey:  apiKey,
		model:   model,
		baseURL: serverURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func TestProviderName(t *testing.T) {
	p := NewProvider("key", "model", 5*time.Second)
	if p.Name() != "anthropic" {
		t.Fatalf("expected 'anthropic', got %s", p.Name())
	}
}

func TestNewProviderDefaultBaseURL(t *testing.T) {
	p := NewProvider("key", "model", 5*time.Second)
	if p.baseURL != defaultBaseURL {
		t.Fatalf("expected default base URL, got %s", p.baseURL)
	}
}

func TestProviderCompleteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Fatalf("expected x-api-key test-key, got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != apiVersion {
			t.Fatalf("expected anthropic-version %s", apiVersion)
		}

		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.System != "You are helpful" {
			t.Fatalf("expected system prompt, got %s", req.System)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Fatalf("expected 1 user message, got %d", len(req.Messages))
		}

		resp := messagesResponse{
			Model: "claude-sonnet-4-20250514",
			Content: []contentBlock{
				{Type: "text", Text: "Hello from Anthropic"},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 8, OutputTokens: 4},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "test-key", "claude-sonnet-4-20250514")

	resp, err := provider.Complete(context.Background(), ai.Request{
		Prompt:       "Say hello",
		SystemPrompt: "You are helpful",
		MaxTokens:    100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Hello from Anthropic" {
		t.Fatalf("expected 'Hello from Anthropic', got %s", resp.Content)
	}
	if resp.Provider != "anthropic" {
		t.Fatalf("expected provider 'anthropic', got %s", resp.Provider)
	}
	if resp.TokensUsed.TotalTokens != 12 {
		t.Fatalf("expected 12 total tokens, got %d", resp.TokensUsed.TotalTokens)
	}
	if resp.Duration <= 0 {
		t.Fatalf("expected positive duration")
	}
}

func TestProviderCompleteDefaultMaxTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req messagesRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.MaxTokens != 1024 {
			t.Fatalf("expected default max_tokens 1024, got %d", req.MaxTokens)
		}

		resp := messagesResponse{
			Model:   "model",
			Content: []contentBlock{{Type: "text", Text: "ok"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "key", "model")
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProviderCompleteAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}{Message: "invalid key"}})
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "bad-key", "model")
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if !errors.Is(err, ai.ErrProviderAuth) {
		t.Fatalf("expected ErrProviderAuth, got %v", err)
	}
}

func TestProviderCompleteRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "key", "model")
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if !errors.Is(err, ai.ErrProviderRateLimit) {
		t.Fatalf("expected ErrProviderRateLimit, got %v", err)
	}
}

func TestProviderCompleteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}{Message: "internal error"}})
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "key", "model")
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error on server error")
	}
}

func TestProviderCompleteInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "key", "model")
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error on invalid JSON")
	}
}

func TestProviderCompleteContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	provider := newTestProvider(server.URL, "key", "model")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := provider.Complete(ctx, ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, ai.ErrProviderTimeout) {
		t.Fatalf("expected ErrProviderTimeout, got %v", err)
	}
}

func TestExtractTextContent(t *testing.T) {
	blocks := []contentBlock{
		{Type: "text", Text: "hello"},
	}
	if got := extractTextContent(blocks); got != "hello" {
		t.Fatalf("expected 'hello', got %s", got)
	}

	if got := extractTextContent(nil); got != "" {
		t.Fatalf("expected empty string for nil blocks, got %s", got)
	}

	nonText := []contentBlock{{Type: "image", Text: ""}}
	if got := extractTextContent(nonText); got != "" {
		t.Fatalf("expected empty string for non-text blocks, got %s", got)
	}
}

func TestMapHTTPErrorSuccessRange(t *testing.T) {
	for _, code := range []int{200, 201, 204} {
		if err := mapHTTPError(code, nil); err != nil {
			t.Fatalf("expected nil for status %d, got %v", code, err)
		}
	}
}

func TestMapHTTPErrorUnparsableBody(t *testing.T) {
	err := mapHTTPError(500, []byte("not json"))
	if err == nil {
		t.Fatalf("expected error for status 500")
	}
}
