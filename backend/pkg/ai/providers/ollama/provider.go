package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"nas-go/api/pkg/ai"
	"net/http"
	"time"
)

type chatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type chatOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	Stream    bool          `json:"stream"`
	Format    string        `json:"format,omitempty"`
	KeepAlive string        `json:"keep_alive,omitempty"`
	Options   *chatOptions  `json:"options,omitempty"`
}

type chatResponse struct {
	Model           string      `json:"model"`
	Message         chatMessage `json:"message"`
	Done            bool        `json:"done"`
	PromptEvalCount int         `json:"prompt_eval_count"`
	EvalCount       int         `json:"eval_count"`
	Error           string      `json:"error"`
}

// Provider implements ai.Provider for a local Ollama daemon.
// Unlike cloud providers it requires no API key and talks to the
// native /api/chat endpoint exposed by the daemon on the LAN/loopback.
type Provider struct {
	baseURL   string
	model     string
	keepAlive string
	client    *http.Client
}

// NewProvider creates an Ollama provider pointing at a running daemon.
// baseURL is the daemon root (e.g. http://localhost:11434), without the API path.
func NewProvider(baseURL, model, keepAlive string, timeout time.Duration) *Provider {
	return &Provider{
		baseURL:   baseURL,
		model:     model,
		keepAlive: keepAlive,
		client:    &http.Client{Timeout: timeout},
	}
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) Complete(ctx context.Context, req ai.Request) (ai.Response, error) {
	start := time.Now()

	body := chatRequest{
		Model:     p.model,
		Messages:  buildMessages(req),
		Stream:    false,
		KeepAlive: p.keepAlive,
		Options:   buildOptions(req),
	}
	if wantsJSON(req.TaskType) {
		body.Format = "json"
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return ai.Response{}, fmt.Errorf("ollama: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return ai.Response{}, fmt.Errorf("ollama: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return ai.Response{}, fmt.Errorf("%w: %v", ai.ErrProviderTimeout, err)
		}
		return ai.Response{}, fmt.Errorf("ollama: request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ai.Response{}, fmt.Errorf("ollama: failed to read response: %w", err)
	}

	if err := mapHTTPError(httpResp.StatusCode, respBody); err != nil {
		return ai.Response{}, err
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return ai.Response{}, fmt.Errorf("ollama: failed to parse response: %w", err)
	}
	if chatResp.Error != "" {
		return ai.Response{}, fmt.Errorf("ollama: model error: %s", chatResp.Error)
	}
	if chatResp.Message.Content == "" {
		return ai.Response{}, fmt.Errorf("ollama: empty response from model")
	}

	return ai.Response{
		Content:  chatResp.Message.Content,
		Model:    chatResp.Model,
		Provider: "ollama",
		TokensUsed: ai.TokenUsage{
			PromptTokens:     chatResp.PromptEvalCount,
			CompletionTokens: chatResp.EvalCount,
			TotalTokens:      chatResp.PromptEvalCount + chatResp.EvalCount,
		},
		Duration: time.Since(start),
	}, nil
}

func buildMessages(req ai.Request) []chatMessage {
	var messages []chatMessage
	if req.SystemPrompt != "" {
		messages = append(messages, chatMessage{Role: "system", Content: req.SystemPrompt})
	}
	// Ollama's /api/chat accepts base64 images on the user message for
	// multimodal models (e.g. gemma3, llava); text-only models ignore them.
	messages = append(messages, chatMessage{Role: "user", Content: req.Prompt, Images: req.Images})
	return messages
}

func buildOptions(req ai.Request) *chatOptions {
	if req.Temperature == 0 && req.MaxTokens == 0 {
		return nil
	}
	return &chatOptions{
		Temperature: req.Temperature,
		NumPredict:  req.MaxTokens,
	}
}

// wantsJSON enables Ollama's structured output mode for task types whose
// callers parse the response as JSON. Local models are far more reliable
// at emitting valid JSON when constrained this way.
func wantsJSON(taskType ai.TaskType) bool {
	switch taskType {
	case ai.TaskExtraction, ai.TaskClassification:
		return true
	default:
		return false
	}
}

func mapHTTPError(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("ollama: API error (status %d): %s", statusCode, errResp.Error)
	}
	return fmt.Errorf("ollama: API error (status %d)", statusCode)
}
