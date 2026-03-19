package ai

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv("AI_OPENAI_API_KEY")
	os.Unsetenv("AI_OPENAI_MODEL")
	os.Unsetenv("AI_OPENAI_BASE_URL")
	os.Unsetenv("AI_ANTHROPIC_API_KEY")
	os.Unsetenv("AI_ANTHROPIC_MODEL")
	os.Unsetenv("AI_TIMEOUT_SECONDS")
	os.Unsetenv("AI_MAX_RETRIES")
	os.Unsetenv("AI_RETRY_BACKOFF_MS")

	cfg := LoadConfig()

	if cfg.OpenAIModel != "gpt-4o-mini" {
		t.Fatalf("expected default model gpt-4o-mini, got %s", cfg.OpenAIModel)
	}
	if cfg.OpenAIBaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected default OpenAI base URL, got %s", cfg.OpenAIBaseURL)
	}
	if cfg.AnthropicModel != "claude-sonnet-4-20250514" {
		t.Fatalf("expected default Anthropic model, got %s", cfg.AnthropicModel)
	}
	if cfg.DefaultTimeout.Seconds() != 30 {
		t.Fatalf("expected 30s timeout, got %v", cfg.DefaultTimeout)
	}
	if cfg.MaxRetries != 2 {
		t.Fatalf("expected 2 retries, got %d", cfg.MaxRetries)
	}
	if cfg.RetryBackoffMS != 500 {
		t.Fatalf("expected 500ms backoff, got %d", cfg.RetryBackoffMS)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("AI_OPENAI_API_KEY", "sk-test")
	t.Setenv("AI_OPENAI_MODEL", "gpt-4o")
	t.Setenv("AI_ANTHROPIC_API_KEY", "ant-test")
	t.Setenv("AI_TIMEOUT_SECONDS", "60")
	t.Setenv("AI_MAX_RETRIES", "5")

	cfg := LoadConfig()

	if cfg.OpenAIAPIKey != "sk-test" {
		t.Fatalf("expected sk-test, got %s", cfg.OpenAIAPIKey)
	}
	if cfg.OpenAIModel != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", cfg.OpenAIModel)
	}
	if cfg.AnthropicAPIKey != "ant-test" {
		t.Fatalf("expected ant-test, got %s", cfg.AnthropicAPIKey)
	}
	if cfg.DefaultTimeout.Seconds() != 60 {
		t.Fatalf("expected 60s timeout, got %v", cfg.DefaultTimeout)
	}
	if cfg.MaxRetries != 5 {
		t.Fatalf("expected 5 retries, got %d", cfg.MaxRetries)
	}
}

func TestLoadConfigInvalidInt(t *testing.T) {
	t.Setenv("AI_MAX_RETRIES", "invalid")
	t.Setenv("AI_TIMEOUT_SECONDS", "-5")

	cfg := LoadConfig()

	if cfg.MaxRetries != 2 {
		t.Fatalf("expected fallback 2 for invalid int, got %d", cfg.MaxRetries)
	}
	if cfg.DefaultTimeout.Seconds() != 30 {
		t.Fatalf("expected fallback 30s for negative int, got %v", cfg.DefaultTimeout)
	}
}
