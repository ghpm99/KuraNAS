package ai

import "time"

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
}

// Response is the standardized output from AI operations,
// including traceability of which model/provider was used.
type Response struct {
	Content    string
	Model      string
	Provider   string
	TokensUsed TokenUsage
	Duration   time.Duration
}

// TokenUsage tracks token consumption for cost/observability purposes.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
