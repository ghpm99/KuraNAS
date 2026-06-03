package aiproviders

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/ai"
)

var (
	ErrInvalidProvider  = errors.New("invalid provider name")
	ErrProviderNotFound = errors.New("provider not found")
)

// Service owns the persisted provider configuration. The API key itself is
// never persisted: it is read from the environment (via ai.Config) only to
// report whether a cloud provider can be enabled.
type Service struct {
	repository RepositoryInterface
	config     ai.Config
	onChange   func()
}

func NewService(repository RepositoryInterface, config ai.Config) *Service {
	return &Service{repository: repository, config: config}
}

// SetOnChange registers a callback invoked after a provider is updated, used
// by the wiring layer to rebuild the active AI service without a restart.
func (s *Service) SetOnChange(fn func()) {
	s.onChange = fn
}

func (s *Service) apiKeyConfigured(name ProviderName) bool {
	switch name {
	case ProviderOpenAI:
		return s.config.OpenAIAPIKey != ""
	case ProviderAnthropic:
		return s.config.AnthropicAPIKey != ""
	default:
		return true // local providers need no key
	}
}

func (s *Service) GetProviders() ([]ProviderDto, error) {
	models, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}

	dtos := make([]ProviderDto, 0, len(models))
	for _, m := range models {
		dtos = append(dtos, m.toDto(s.apiKeyConfigured(m.Name)))
	}
	return dtos, nil
}

func (s *Service) GetProviderModels() ([]ProviderModel, error) {
	return s.repository.GetAll()
}

func (s *Service) UpdateProvider(name ProviderName, dto UpdateProviderDto) (ProviderDto, error) {
	if !name.IsValid() {
		return ProviderDto{}, ErrInvalidProvider
	}

	existing, err := s.repository.GetByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProviderDto{}, ErrProviderNotFound
		}
		return ProviderDto{}, err
	}

	saved, err := s.repository.Update(dto.applyTo(existing))
	if err != nil {
		return ProviderDto{}, err
	}

	if s.onChange != nil {
		s.onChange()
	}

	return saved.toDto(s.apiKeyConfigured(saved.Name)), nil
}

// EnsureDefaults seeds the three known providers from the current environment
// configuration on first boot. It is idempotent: existing rows are untouched.
func (s *Service) EnsureDefaults() error {
	for _, model := range s.defaultModels() {
		if err := s.repository.InsertIfAbsent(model); err != nil {
			return fmt.Errorf("ensure default %s: %w", model.Name, err)
		}
	}
	return nil
}

// defaultOllamaTimeoutSeconds is intentionally higher than the cloud default:
// local inference (especially the first request after a model is loaded) is
// considerably slower than a hosted API.
const defaultOllamaTimeoutSeconds = 120

func (s *Service) defaultModels() []ProviderModel {
	cloudTimeout := int(s.config.DefaultTimeout.Seconds())
	if cloudTimeout <= 0 {
		cloudTimeout = 30
	}

	baseParams := func(timeoutSeconds int) ProviderParams {
		return ProviderParams{
			TimeoutSeconds: timeoutSeconds,
			MaxRetries:     s.config.MaxRetries,
			RetryBackoffMS: s.config.RetryBackoffMS,
		}
	}

	ollamaParams := baseParams(defaultOllamaTimeoutSeconds)
	ollamaParams.KeepAlive = s.config.OllamaKeepAlive

	return []ProviderModel{
		{
			Name:     ProviderOllama,
			Enabled:  s.config.OllamaEnabled,
			Model:    s.config.OllamaModel,
			BaseURL:  s.config.OllamaBaseURL,
			Priority: 0,
			Params:   ollamaParams,
		},
		{
			Name:     ProviderOpenAI,
			Enabled:  s.config.OpenAIAPIKey != "",
			Model:    s.config.OpenAIModel,
			BaseURL:  s.config.OpenAIBaseURL,
			Priority: 1,
			Params:   baseParams(cloudTimeout),
		},
		{
			Name:     ProviderAnthropic,
			Enabled:  s.config.AnthropicAPIKey != "",
			Model:    s.config.AnthropicModel,
			BaseURL:  "",
			Priority: 2,
			Params:   baseParams(cloudTimeout),
		},
	}
}
