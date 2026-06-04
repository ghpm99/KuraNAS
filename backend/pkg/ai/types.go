package ai

import (
	"encoding/json"
	"time"
)

// TaskType determines which provider/model strategy to use for a given request.
type TaskType string

const (
	TaskGeneration     TaskType = "generation"
	TaskSummarization  TaskType = "summarization"
	TaskClassification TaskType = "classification"
	TaskExtraction     TaskType = "extraction"
	TaskSimple         TaskType = "simple"
	TaskComplex        TaskType = "complex"
)

// Request is the standardized input for all AI operations,
// independent of the underlying provider.
type Request struct {
	TaskType     TaskType
	Prompt       string
	SystemPrompt string
	MaxTokens    int
	Temperature  float64
	Metadata     map[string]string
	// Images holds base64-encoded image data for multimodal (vision) requests.
	// Providers that support vision attach them to the user message; others
	// ignore them and fall back to a text-only request.
	Images []string
	// Tools, when set, are offered to the model for function calling. Providers
	// that support tools advertise them and surface any requests in
	// Response.ToolCalls; providers that do not simply ignore the field.
	Tools []ToolDefinition
}

// ToolDefinition describes a callable function exposed to the model. Parameters
// is a JSON Schema object describing the arguments.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolCall is a model's request to invoke a tool with the given JSON arguments.
type ToolCall struct {
	Name      string
	Arguments json.RawMessage
}

// Response is the standardized output from AI operations,
// including traceability of which model/provider was used.
type Response struct {
	Content    string
	Model      string
	Provider   string
	TokensUsed TokenUsage
	Duration   time.Duration
	// ToolCalls holds any function-call requests the model made instead of (or
	// alongside) a textual answer.
	ToolCalls []ToolCall
}

// TokenUsage tracks token consumption for cost/observability purposes.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
