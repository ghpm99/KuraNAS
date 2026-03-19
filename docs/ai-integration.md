# AI Integration Architecture

## Overview

The `pkg/ai` package provides a provider-agnostic AI/LLM integration layer for KuraNAS. It uses the **Strategy Pattern** to route different task types to different AI models/providers, with built-in retry, fallback, and standardized request/response types.

## Package Structure

```
backend/pkg/ai/
├── interfaces.go              # Provider, ServiceInterface contracts
├── types.go                   # Request, Response, TaskType, TokenUsage
├── errors.go                  # Domain-specific errors
├── config.go                  # Config via env vars (AI_*)
├── router.go                  # Router with Strategy Pattern
├── service.go                 # Service with retry + fallback orchestration
└── providers/
    ├── openai/
    │   └── provider.go        # OpenAI chat completions via net/http
    └── anthropic/
        └── provider.go        # Anthropic Messages API via net/http
```

## Architectural Decisions

1. **Location**: `pkg/ai/` — shared infrastructure, same level as `pkg/logger/`
2. **No external SDK**: uses `net/http` directly, consistent with the project's approach
3. **Strategy Pattern**: `Router` maps `TaskType` → `Provider` for dynamic selection
4. **Interface-first**: `Provider` and `ServiceInterface` as contracts for testability
5. **Config via env vars**: `AI_*` prefix, integrated with existing config pattern
6. **Retry with backoff**: built into the service, no external library
7. **Fallback**: support for fallback provider per task type

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AI_OPENAI_API_KEY` | — | OpenAI API key |
| `AI_OPENAI_MODEL` | `gpt-4o-mini` | OpenAI model identifier |
| `AI_OPENAI_BASE_URL` | `https://api.openai.com/v1` | OpenAI API base URL |
| `AI_ANTHROPIC_API_KEY` | — | Anthropic API key |
| `AI_ANTHROPIC_MODEL` | `claude-sonnet-4-20250514` | Anthropic model identifier |
| `AI_TIMEOUT_SECONDS` | `30` | Request timeout in seconds |
| `AI_MAX_RETRIES` | `2` | Max retry attempts on retryable errors |
| `AI_RETRY_BACKOFF_MS` | `500` | Base backoff between retries in ms |

## Core Concepts

### TaskType

Determines which provider/model strategy to use:

```go
ai.TaskGeneration     // text generation
ai.TaskSummarization  // summarization
ai.TaskClassification // classification
ai.TaskExtraction     // structured extraction
ai.TaskSimple         // cheap, fast models
ai.TaskComplex        // expensive, powerful models
```

### Provider Interface

Each provider implements this contract:

```go
type Provider interface {
    Complete(ctx context.Context, req Request) (Response, error)
    Name() string
}
```

### Router (Strategy Pattern)

Maps task types to providers with optional fallback:

```go
router := ai.NewRouter()
router.Register(ai.TaskSimple, openaiProvider)
router.RegisterWithFallback(ai.TaskComplex, anthropicProvider, openaiProvider)
```

### Service

Orchestrates requests: resolves provider via router, handles retry and fallback:

```go
service := ai.NewService(router, cfg)
resp, err := service.Execute(ctx, ai.Request{
    TaskType: ai.TaskSummarization,
    Prompt:   "Summarize this...",
})
```

## Usage Example

```go
package main

import (
    "context"
    "fmt"
    "nas-go/api/pkg/ai"
    "nas-go/api/pkg/ai/providers/openai"
    "nas-go/api/pkg/ai/providers/anthropic"
)

func example() {
    cfg := ai.LoadConfig()

    // Create providers
    openaiProvider := openai.NewProvider(
        cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.OpenAIBaseURL, cfg.DefaultTimeout,
    )
    anthropicProvider := anthropic.NewProvider(
        cfg.AnthropicAPIKey, cfg.AnthropicModel, cfg.DefaultTimeout,
    )

    // Configure router (Strategy Pattern)
    router := ai.NewRouter()
    router.Register(ai.TaskSimple, openaiProvider)
    router.Register(ai.TaskClassification, openaiProvider)
    router.Register(ai.TaskSummarization, openaiProvider)
    router.RegisterWithFallback(ai.TaskComplex, anthropicProvider, openaiProvider)

    // Create service
    aiService := ai.NewService(router, cfg)

    // Execute
    resp, err := aiService.Execute(context.Background(), ai.Request{
        TaskType:     ai.TaskSummarization,
        Prompt:       "Summarize this article: ...",
        SystemPrompt: "You are a summarization assistant.",
        MaxTokens:    500,
        Temperature:  0.3,
        Metadata:     map[string]string{"source": "diary"},
    })
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Response from %s (%s): %s\n", resp.Provider, resp.Model, resp.Content)
    fmt.Printf("Tokens used: %d (duration: %v)\n", resp.TokensUsed.TotalTokens, resp.Duration)
}
```

### Integration with `context.go`

When ready to wire into the application:

```go
// In internal/app/context.go
func newAIService() ai.ServiceInterface {
    cfg := ai.LoadConfig()
    openaiProvider := openai.NewProvider(
        cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.OpenAIBaseURL, cfg.DefaultTimeout,
    )
    anthropicProvider := anthropic.NewProvider(
        cfg.AnthropicAPIKey, cfg.AnthropicModel, cfg.DefaultTimeout,
    )

    router := ai.NewRouter()
    router.Register(ai.TaskSimple, openaiProvider)
    router.RegisterWithFallback(ai.TaskComplex, anthropicProvider, openaiProvider)

    return ai.NewService(router, cfg)
}
```

## Adding a New Provider

To add a new provider (e.g., Google Gemini):

1. Create `backend/pkg/ai/providers/gemini/provider.go`
2. Implement the `ai.Provider` interface:

```go
package gemini

type Provider struct { /* fields */ }

func NewProvider(apiKey, model string, timeout time.Duration) *Provider { /* ... */ }
func (p *Provider) Name() string { return "gemini" }
func (p *Provider) Complete(ctx context.Context, req ai.Request) (ai.Response, error) { /* ... */ }
```

3. Register in the router:

```go
router.Register(ai.TaskExtraction, geminiProvider)
```

4. Add env vars to `config.go` if needed
5. Create tests using `httptest.NewServer`

**No changes needed to `Service`, `Router`, or existing types.**

## Error Handling

The package defines domain-specific errors:

| Error | Description | Retryable |
|-------|-------------|-----------|
| `ErrNoProviderForTask` | No provider configured for the task type | No |
| `ErrEmptyPrompt` | Request prompt is empty | No |
| `ErrProviderTimeout` | Provider request timed out | Yes |
| `ErrProviderRateLimit` | Provider rate limit exceeded | Yes |
| `ErrProviderAuth` | Provider authentication failed | No |
| `ErrAllAttemptsFailed` | All retry attempts exhausted | No |

## Retry & Fallback Behavior

1. On retryable errors (`Timeout`, `RateLimit`), the service retries up to `MaxRetries` times with exponential backoff
2. On non-retryable errors (`Auth`, unknown), it fails immediately
3. If a fallback provider is configured and the primary fails (any error), the fallback is tried with the same retry logic
4. Context cancellation is respected between retries

## Future Evolution

| Feature | Where to Add |
|---------|-------------|
| **Response caching** | Decorator around `Provider` or middleware in `Service` |
| **Observability/Metrics** | Interceptor in `Service.Execute` (log duration, tokens, provider) |
| **Circuit Breaker** | Wrapper around `Provider` with failure counters |
| **Feature Flags** | Conditional in `Router.Resolve` before selecting provider |
| **Prompt Versioning** | `PromptVersion` field in `Request.Metadata` |
| **Background Jobs** | Integrate with existing `worker.JobScheduler` via new step type |
| **Streaming** | `Stream(ctx, req) (<-chan string, error)` method on `Provider` |
| **History Persistence** | Repository + Model to save Request/Response in DB |

## Test Coverage

- **41 tests**, all passing
- **Coverage**: 100% (core), ~91% (providers)
- **Zero external dependencies** added
- Tests use `httptest.NewServer` for provider tests and function-based mocks for service/router tests, consistent with project patterns
