package openai

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
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// Provider implements ai.Provider for the OpenAI API.
type Provider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewProvider creates an OpenAI provider with the given configuration.
func NewProvider(apiKey, model, baseURL string, timeout time.Duration) *Provider {
	return &Provider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) Complete(ctx context.Context, req ai.Request) (ai.Response, error) {
	start := time.Now()

	messages := buildMessages(req)
	body := chatRequest{
		Model:       p.model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return ai.Response{}, fmt.Errorf("openai: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return ai.Response{}, fmt.Errorf("openai: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return ai.Response{}, fmt.Errorf("%w: %v", ai.ErrProviderTimeout, err)
		}
		return ai.Response{}, fmt.Errorf("openai: request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ai.Response{}, fmt.Errorf("openai: failed to read response: %w", err)
	}

	if err := mapHTTPError(httpResp.StatusCode, respBody); err != nil {
		return ai.Response{}, err
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return ai.Response{}, fmt.Errorf("openai: failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return ai.Response{}, fmt.Errorf("openai: empty response from model")
	}

	return ai.Response{
		Content:  chatResp.Choices[0].Message.Content,
		Model:    chatResp.Model,
		Provider: "openai",
		TokensUsed: ai.TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
		Duration: time.Since(start),
	}, nil
}

func buildMessages(req ai.Request) []chatMessage {
	var messages []chatMessage
	if req.SystemPrompt != "" {
		messages = append(messages, chatMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, chatMessage{Role: "user", Content: req.Prompt})
	return messages
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
			return fmt.Errorf("openai: API error (status %d): %s", statusCode, errResp.Error.Message)
		}
		return fmt.Errorf("openai: API error (status %d)", statusCode)
	}
}
