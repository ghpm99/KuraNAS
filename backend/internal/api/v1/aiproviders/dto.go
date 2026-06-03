package aiproviders

// ProviderDto is the API representation of a provider configuration. It never
// exposes the API key; it only reports whether one is configured in the env.
type ProviderDto struct {
	Name             string         `json:"name"`
	Enabled          bool           `json:"enabled"`
	Model            string         `json:"model"`
	BaseURL          string         `json:"base_url"`
	Priority         int            `json:"priority"`
	Params           ProviderParams `json:"params"`
	RequiresAPIKey   bool           `json:"requires_api_key"`
	APIKeyConfigured bool           `json:"api_key_configured"`
}

// UpdateProviderDto is the editable subset of a provider configuration.
type UpdateProviderDto struct {
	Enabled  bool           `json:"enabled"`
	Model    string         `json:"model"`
	BaseURL  string         `json:"base_url"`
	Priority int            `json:"priority"`
	Params   ProviderParams `json:"params"`
}

func (m ProviderModel) toDto(apiKeyConfigured bool) ProviderDto {
	return ProviderDto{
		Name:             string(m.Name),
		Enabled:          m.Enabled,
		Model:            m.Model,
		BaseURL:          m.BaseURL,
		Priority:         m.Priority,
		Params:           m.Params,
		RequiresAPIKey:   m.Name.RequiresAPIKey(),
		APIKeyConfigured: apiKeyConfigured,
	}
}

func (d UpdateProviderDto) applyTo(model ProviderModel) ProviderModel {
	model.Enabled = d.Enabled
	model.Model = d.Model
	model.BaseURL = d.BaseURL
	model.Priority = d.Priority
	model.Params = d.Params
	return model
}
