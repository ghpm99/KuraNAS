package app

import (
	"testing"

	"nas-go/api/internal/api/v1/aiproviders"
	"nas-go/api/pkg/ai"
)

func TestBuildProvidersFromModelsNamedMap(t *testing.T) {
	cfg := ai.Config{OpenAIAPIKey: "k", OllamaKeepAlive: "5m"}
	models := []aiproviders.ProviderModel{
		{Name: aiproviders.ProviderOllama, Enabled: true, Model: "llama3", BaseURL: "http://localhost:11434"},
		{Name: aiproviders.ProviderOpenAI, Enabled: true, Model: "gpt-4o-mini"},
		{Name: aiproviders.ProviderAnthropic, Enabled: true, Model: "claude"}, // no key -> skipped
	}

	providers, named := buildProvidersFromModels(models, cfg)

	// ollama + openai enabled; anthropic skipped for the missing key.
	if len(providers) != 2 {
		t.Fatalf("providers = %d, want 2 (ollama + openai)", len(providers))
	}
	if named[string(aiproviders.ProviderOllama)] == nil {
		t.Fatal("ollama must be in the named map")
	}
	if named[string(aiproviders.ProviderOpenAI)] == nil {
		t.Fatal("openai must be in the named map")
	}
	if named[string(aiproviders.ProviderAnthropic)] != nil {
		t.Fatal("anthropic must be absent (no API key)")
	}
}

func TestBuildServiceFromProvidersEmptyIsNil(t *testing.T) {
	if buildServiceFromProviders(nil) != nil {
		t.Fatal("expected nil service when no providers are enabled")
	}
}

func TestBuildProvidersFromModelsSkipsDisabled(t *testing.T) {
	cfg := ai.Config{OllamaKeepAlive: "5m"}
	models := []aiproviders.ProviderModel{
		{Name: aiproviders.ProviderOllama, Enabled: false, Model: "llama3", BaseURL: "http://localhost:11434"},
	}
	providers, named := buildProvidersFromModels(models, cfg)
	if len(providers) != 0 || len(named) != 0 {
		t.Fatalf("disabled provider must be excluded, got %d providers / %d named", len(providers), len(named))
	}
	if svc := buildServiceFromProviders(providers); svc != nil {
		t.Fatal("expected nil service for empty provider list")
	}
}
