package ollama

// PullStepPayload is the worker step payload for a model download. It carries
// the daemon base URL resolved at enqueue time so the worker does not need to
// re-resolve provider configuration.
type PullStepPayload struct {
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
}
