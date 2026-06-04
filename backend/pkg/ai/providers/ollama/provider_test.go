package ollama

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
	p := NewProvider("http://localhost:11434", "llama3.1", "5m", 5*time.Second)
	if p.Name() != "ollama" {
		t.Fatalf("expected 'ollama', got %s", p.Name())
	}
}

func TestProviderForwardsImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		userMsg := req.Messages[len(req.Messages)-1]
		if len(userMsg.Images) != 1 || userMsg.Images[0] != "base64data" {
			t.Fatalf("expected image forwarded on user message, got %+v", userMsg.Images)
		}
		json.NewEncoder(w).Encode(chatResponse{
			Model:   "gemma3",
			Message: chatMessage{Role: "assistant", Content: "ok"},
			Done:    true,
		})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "gemma3", "5m", 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{
		TaskType: ai.TaskClassification,
		Prompt:   "describe",
		Images:   []string{"base64data"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderCompleteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("expected /api/chat, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type")
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model != "llama3.1" {
			t.Fatalf("expected model llama3.1, got %s", req.Model)
		}
		if req.Stream {
			t.Fatalf("expected stream=false")
		}
		if req.KeepAlive != "5m" {
			t.Fatalf("expected keep_alive 5m, got %s", req.KeepAlive)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages (system + user), got %d", len(req.Messages))
		}

		resp := chatResponse{
			Model:           "llama3.1",
			Message:         chatMessage{Role: "assistant", Content: "Hello from Ollama"},
			Done:            true,
			PromptEvalCount: 12,
			EvalCount:       8,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)

	resp, err := provider.Complete(context.Background(), ai.Request{
		Prompt:       "Say hello",
		SystemPrompt: "You are helpful",
		MaxTokens:    100,
		Temperature:  0.7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Hello from Ollama" {
		t.Fatalf("expected 'Hello from Ollama', got %s", resp.Content)
	}
	if resp.Provider != "ollama" {
		t.Fatalf("expected provider 'ollama', got %s", resp.Provider)
	}
	if resp.TokensUsed.TotalTokens != 20 {
		t.Fatalf("expected 20 total tokens, got %d", resp.TokensUsed.TotalTokens)
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
		json.NewEncoder(w).Encode(chatResponse{
			Model:   "llama3.1",
			Message: chatMessage{Role: "assistant", Content: "response"},
			Done:    true,
		})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProviderCompleteSetsJSONFormatForStructuredTasks(t *testing.T) {
	for _, taskType := range []ai.TaskType{ai.TaskExtraction, ai.TaskClassification} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req chatRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Format != "json" {
				t.Fatalf("expected format=json for task %s, got %q", taskType, req.Format)
			}
			json.NewEncoder(w).Encode(chatResponse{
				Model:   "llama3.1",
				Message: chatMessage{Content: "{}"},
				Done:    true,
			})
		}))

		provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
		_, err := provider.Complete(context.Background(), ai.Request{TaskType: taskType, Prompt: "test"})
		server.Close()
		if err != nil {
			t.Fatalf("expected no error for task %s, got %v", taskType, err)
		}
	}
}

func TestProviderCompleteNoJSONFormatForGeneration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Format != "" {
			t.Fatalf("expected no format for generation task, got %q", req.Format)
		}
		json.NewEncoder(w).Encode(chatResponse{
			Model:   "llama3.1",
			Message: chatMessage{Content: "free text"},
			Done:    true,
		})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{TaskType: ai.TaskGeneration, Prompt: "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProviderCompleteModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "model 'missing' not found"})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "missing", "5m", 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error on model not found")
	}
}

func TestProviderCompleteErrorInBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{Error: "something broke"})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error when response carries an error field")
	}
}

func TestProviderCompleteEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{Model: "llama3.1", Done: true})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	_, err := provider.Complete(context.Background(), ai.Request{Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error on empty content")
	}
}

func TestProviderCompleteInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
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

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := provider.Complete(ctx, ai.Request{Prompt: "test"})
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

func TestBuildOptions(t *testing.T) {
	if buildOptions(ai.Request{}) != nil {
		t.Fatalf("expected nil options when temperature and max tokens are zero")
	}
	opts := buildOptions(ai.Request{Temperature: 0.5, MaxTokens: 42})
	if opts == nil || opts.Temperature != 0.5 || opts.NumPredict != 42 {
		t.Fatalf("expected options carrying temperature and num_predict, got %+v", opts)
	}
}

func TestWantsJSON(t *testing.T) {
	for _, tt := range []ai.TaskType{ai.TaskExtraction, ai.TaskClassification} {
		if !wantsJSON(tt) {
			t.Fatalf("expected wantsJSON true for %s", tt)
		}
	}
	for _, tt := range []ai.TaskType{ai.TaskGeneration, ai.TaskSummarization, ai.TaskSimple, ai.TaskComplex} {
		if wantsJSON(tt) {
			t.Fatalf("expected wantsJSON false for %s", tt)
		}
	}
}

func TestMapHTTPError(t *testing.T) {
	if err := mapHTTPError(200, nil); err != nil {
		t.Fatalf("expected nil for 200, got %v", err)
	}
	if err := mapHTTPError(500, []byte(`{"error":"boom"}`)); err == nil {
		t.Fatalf("expected error for 500 with body")
	}
	if err := mapHTTPError(500, []byte("not json")); err == nil {
		t.Fatalf("expected error for 500 with unparsable body")
	}
}

func TestCompleteStreamSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if !req.Stream {
			t.Fatalf("expected stream=true")
		}
		enc := json.NewEncoder(w)
		enc.Encode(chatResponse{Model: "llama3.1", Message: chatMessage{Role: "assistant", Content: "Olá"}})
		enc.Encode(chatResponse{Model: "llama3.1", Message: chatMessage{Role: "assistant", Content: ", mundo"}})
		enc.Encode(chatResponse{Model: "llama3.1", Done: true, PromptEvalCount: 3, EvalCount: 5})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)

	var chunks []string
	resp, err := provider.CompleteStream(context.Background(), ai.Request{TaskType: ai.TaskGeneration, Prompt: "oi"}, func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 || chunks[0] != "Olá" || chunks[1] != ", mundo" {
		t.Fatalf("unexpected chunks: %v", chunks)
	}
	if resp.Content != "Olá, mundo" {
		t.Fatalf("expected accumulated content, got %q", resp.Content)
	}
	if resp.TokensUsed.TotalTokens != 8 {
		t.Fatalf("expected token totals, got %+v", resp.TokensUsed)
	}
}

func TestCompleteStreamHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"boom"}`))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	_, err := provider.CompleteStream(context.Background(), ai.Request{TaskType: ai.TaskGeneration, Prompt: "oi"}, func(string) error { return nil })
	if err == nil {
		t.Fatalf("expected error for non-2xx status")
	}
}

func TestCompleteStreamModelError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{Error: "model exploded"})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	_, err := provider.CompleteStream(context.Background(), ai.Request{TaskType: ai.TaskGeneration, Prompt: "oi"}, func(string) error { return nil })
	if err == nil {
		t.Fatalf("expected model error")
	}
}

func TestCompleteStreamCallbackError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{Message: chatMessage{Content: "hi"}})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	sentinel := errors.New("stop")
	_, err := provider.CompleteStream(context.Background(), ai.Request{TaskType: ai.TaskGeneration, Prompt: "oi"}, func(string) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected callback error to propagate, got %v", err)
	}
}

func TestCompleteWithToolsParsesToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Tools) != 1 || req.Tools[0].Function.Name != "buscar" {
			t.Fatalf("expected tool advertised, got %+v", req.Tools)
		}
		if req.Format == "json" {
			t.Fatalf("format json must be disabled when tools are present")
		}
		json.NewEncoder(w).Encode(chatResponse{
			Model: "llama3.1",
			Message: chatMessage{
				Role: "assistant",
				ToolCalls: []chatToolCall{
					{Function: chatToolCallFunction{Name: "buscar", Arguments: json.RawMessage(`{"query":"ipva"}`)}},
				},
			},
			Done: true,
		})
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "llama3.1", "5m", 5*time.Second)
	resp, err := provider.Complete(context.Background(), ai.Request{
		TaskType: ai.TaskClassification, // would normally request json format
		Prompt:   "procura ipva",
		Tools: []ai.ToolDefinition{
			{Name: "buscar", Description: "busca", Parameters: json.RawMessage(`{"type":"object"}`)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "buscar" {
		t.Fatalf("expected parsed tool call, got %+v", resp.ToolCalls)
	}
}
