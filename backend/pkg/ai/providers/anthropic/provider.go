package anthropic

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

const defaultBaseURL = "https://api.anthropic.com/v1"
const apiVersion = "2023-06-01"

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type messagesRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    string         `json:"system,omitempty"`
	Messages  []messageBlock `json:"messages"`
}

type messageBlock struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []contentBlock `json:"content"`
	Model   string         `json:"model"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type errorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Provider implements ai.Provider for the Anthropic Messages API.
type Provider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewProvider creates an Anthropic provider with the given configuration.
func NewProvider(apiKey, model string, timeout time.Duration) *Provider {
	return &Provider{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *Provider) Name() string {
	return "anthropic"
}

func (p *Provider) Complete(ctx context.Context, req ai.Request) (ai.Response, error) {
	start := time.Now()

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	body := messagesRequest{
		Model:     p.model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages:  []messageBlock{{Role: "user", Content: req.Prompt}},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return ai.Response{}, fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return ai.Response{}, fmt.Errorf("anthropic: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return ai.Response{}, fmt.Errorf("%w: %v", ai.ErrProviderTimeout, err)
		}
		return ai.Response{}, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ai.Response{}, fmt.Errorf("anthropic: failed to read response: %w", err)
	}

	if err := mapHTTPError(httpResp.StatusCode, respBody); err != nil {
		return ai.Response{}, err
	}

	var msgResp messagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return ai.Response{}, fmt.Errorf("anthropic: failed to parse response: %w", err)
	}

	content := extractTextContent(msgResp.Content)

	return ai.Response{
		Content:  content,
		Model:    msgResp.Model,
		Provider: "anthropic",
		TokensUsed: ai.TokenUsage{
			PromptTokens:     msgResp.Usage.InputTokens,
			CompletionTokens: msgResp.Usage.OutputTokens,
			TotalTokens:      msgResp.Usage.InputTokens + msgResp.Usage.OutputTokens,
		},
		Duration: time.Since(start),
	}, nil
}

func extractTextContent(blocks []contentBlock) string {
	for _, b := range blocks {
		if b.Type == "text" {
			return b.Text
		}
	}
	return ""
}

func mapHTTPError(statusCode int, body []byte) error {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return nil
	case statusCode == 401:
		return fmt.Errorf("%w: invalid API key", ai.ErrProviderAuth)
	case statusCode == 429:
		return ai.ErrProviderRateLimit
	default:
		var errResp errorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("anthropic: API error (status %d): %s", statusCode, errResp.Error.Message)
		}
		return fmt.Errorf("anthropic: API error (status %d)", statusCode)
	}
}
