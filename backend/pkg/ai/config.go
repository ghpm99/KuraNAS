package ai

import (
	"os"
	"strconv"
	"time"
)

// Config holds all AI-related configuration loaded from environment variables.
type Config struct {
	OpenAIAPIKey    string
	OpenAIModel     string
	OpenAIBaseURL   string
	AnthropicAPIKey string
	AnthropicModel  string
	OllamaEnabled   bool
	OllamaBaseURL   string
	OllamaModel     string
	OllamaKeepAlive string
	// OllamaAutostart controls whether the backend tries to spawn `ollama serve`
	// at boot when the daemon is unreachable and the provider is enabled.
	OllamaAutostart bool
	// OllamaBinary is the daemon binary name/path used for the autostart spawn.
	OllamaBinary string
	// OllamaStartTimeout caps how long boot waits for a freshly spawned daemon to
	// answer before giving up (the server keeps booting regardless).
	OllamaStartTimeout time.Duration
	DefaultTimeout     time.Duration
	MaxRetries         int
	RetryBackoffMS     int
}

// LoadConfig reads AI configuration from environment variables.
func LoadConfig() Config {
	return Config{
		OpenAIAPIKey:    os.Getenv("AI_OPENAI_API_KEY"),
		OpenAIModel:     envOrDefault("AI_OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:   envOrDefault("AI_OPENAI_BASE_URL", "https://api.openai.com/v1"),
		AnthropicAPIKey: os.Getenv("AI_ANTHROPIC_API_KEY"),
		AnthropicModel:  envOrDefault("AI_ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
		OllamaEnabled:   envBoolOrDefault("AI_OLLAMA_ENABLED", false),
		OllamaBaseURL:   envOrDefault("AI_OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:     envOrDefault("AI_OLLAMA_MODEL", "llama3.1"),
		OllamaKeepAlive: envOrDefault("AI_OLLAMA_KEEP_ALIVE", "5m"),
		OllamaAutostart: envBoolOrDefault("AI_OLLAMA_AUTOSTART", true),
		OllamaBinary:    envOrDefault("AI_OLLAMA_BIN", "ollama"),
		OllamaStartTimeout: time.Duration(
			envIntOrDefault("AI_OLLAMA_START_TIMEOUT_SECONDS", 30),
		) * time.Second,
		DefaultTimeout: time.Duration(envIntOrDefault("AI_TIMEOUT_SECONDS", 30)) * time.Second,
		MaxRetries:     envIntOrDefault("AI_MAX_RETRIES", 2),
		RetryBackoffMS: envIntOrDefault("AI_RETRY_BACKOFF_MS", 500),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envBoolOrDefault(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
