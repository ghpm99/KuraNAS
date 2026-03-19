package openai

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

func TestProviderName(t *testing.T) {
	p := NewProvider("key", "model", "http://localhost", 5*time.Second)
	if p.Name() != "openai" {
		t.Fatalf("expected 'openai', got %s", p.Name())
	}
}

func TestProviderCompleteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type")
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model != "gpt-4o-mini" {
			t.Fatalf("expected model gpt-4o-mini, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages (system + user), got %d", len(req.Messages))
		}

		resp := chatResponse{
			Model: "gpt-4o-mini",
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "Hello from OpenAI"}},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewProvider("test-key", "gpt-4o-mini", server.URL, 5*time.Second)

	resp, err := provider.Complete(context.Background(), ai.Request{
		Prompt:       "Say hello",
		SystemPrompt: "You are helpful",
		MaxTokens:    100,
		Temperature:  0.7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Hello from OpenAI" {
		t.Fatalf("expected 'Hello from OpenAI', got %s", resp.Content)
	}
	if resp.Provider != "openai" {
		t.Fatalf("expected provider 'openai', got %s", resp.Provider)
	}
	if resp.TokensUsed.TotalTokens != 15 {
		t.Fatalf("expected 15 total tokens, got %d", resp.TokensUsed.TotalTokens)
	}
	if resp.Duration <= 0 {
		t.Fatalf("expected positive duration")
	}
}

func TestProviderCompleteWithoutSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Messages) != 1 {
			t.Fatalf("expected 1 message (user only), got %d", len(req.Messages))
		}

		resp := chatResponse{
			Model: "gpt-4o-mini",
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "response"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewProvider("key", "gpt-4o-mini", server.URL, 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProviderCompleteAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		}{Message: "invalid key", Type: "auth"}})
	}))
	defer server.Close()

	provider := NewProvider("bad-key", "model", server.URL, 5*time.Second)
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

	provider := NewProvider("key", "model", server.URL, 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if !errors.Is(err, ai.ErrProviderRateLimit) {
		t.Fatalf("expected ErrProviderRateLimit, got %v", err)
	}
}

func TestProviderCompleteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		}{Message: "internal error"}})
	}))
	defer server.Close()

	provider := NewProvider("key", "model", server.URL, 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error on server error")
	}
}

func TestProviderCompleteEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{Model: "model", Choices: nil}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewProvider("key", "model", server.URL, 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error on empty choices")
	}
}

func TestProviderCompleteInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	provider := NewProvider("key", "model", server.URL, 5*time.Second)
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

	provider := NewProvider("key", "model", server.URL, 5*time.Second)
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

func TestBuildMessagesWithSystem(t *testing.T) {
	msgs := buildMessages(ai.Request{SystemPrompt: "sys", Prompt: "user"})
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" || msgs[0].Content != "sys" {
		t.Fatalf("expected system message first")
	}
	if msgs[1].Role != "user" || msgs[1].Content != "user" {
		t.Fatalf("expected user message second")
	}
}

func TestBuildMessagesWithoutSystem(t *testing.T) {
	msgs := buildMessages(ai.Request{Prompt: "user"})
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestMapHTTPErrorSuccessRange(t *testing.T) {
	for _, code := range []int{200, 201, 204} {
		if err := mapHTTPError(code, nil); err != nil {
			t.Fatalf("expected nil for status %d, got %v", code, err)
		}
	}
}

func TestMapHTTPErrorWithUnparsableBody(t *testing.T) {
	err := mapHTTPError(500, []byte("not json"))
	if err == nil {
		t.Fatalf("expected error for status 500")
	}
}
