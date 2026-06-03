package aiproviders

import (
	"encoding/json"
	"time"
)

// ProviderName identifies a supported AI provider. It matches the CHECK
// constraint on the ai_providers table.
type ProviderName string

const (
	ProviderOllama    ProviderName = "ollama"
	ProviderOpenAI    ProviderName = "openai"
	ProviderAnthropic ProviderName = "anthropic"
)

var AllProviderNames = []ProviderName{ProviderOllama, ProviderOpenAI, ProviderAnthropic}

func (n ProviderName) IsValid() bool {
	switch n {
	case ProviderOllama, ProviderOpenAI, ProviderAnthropic:
		return true
	default:
		return false
	}
}

// RequiresAPIKey reports whether the provider needs a secret API key (read
// from the environment, never persisted). Local providers do not.
func (n ProviderName) RequiresAPIKey() bool {
	return n == ProviderOpenAI || n == ProviderAnthropic
}

// ProviderParams holds optional, provider-specific tuning persisted as JSON.
type ProviderParams struct {
	KeepAlive      string `json:"keep_alive,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// ProviderModel mirrors a row in the ai_providers table. The API key is never
// stored here; it stays in the environment.
type ProviderModel struct {
	ID        int
	Name      ProviderName
	Enabled   bool
	Model     string
	BaseURL   string
	Priority  int
	Params    ProviderParams
	CreatedAt time.Time
	UpdatedAt time.Time
}

// encodeParams serialises Params for the JSON column.
func encodeParams(params ProviderParams) ([]byte, error) {
	return json.Marshal(params)
}

// decodeParams parses the JSON column into Params, tolerating null/empty.
func decodeParams(raw []byte) (ProviderParams, error) {
	params := ProviderParams{}
	if len(raw) == 0 {
		return params, nil
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return ProviderParams{}, err
	}
	return params, nil
}
