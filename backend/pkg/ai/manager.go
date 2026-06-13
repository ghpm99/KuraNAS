package ai

import (
	"context"
	"sync"
)

// Manager is a hot-swappable ServiceInterface. Consumers hold a stable
// reference to the Manager while the underlying service (provider router)
// can be rebuilt at runtime — e.g. when providers are enabled/disabled
// through the UI — without restarting the application.
//
// When no providers are enabled the inner service is nil and Execute
// returns ErrServiceUnavailable, which callers already treat as a graceful
// "AI unavailable" signal.
type Manager struct {
	mu    sync.RWMutex
	inner ServiceInterface
	// named exposes each enabled provider by its registry name (e.g. "ollama",
	// "openai", "anthropic) so a caller can pin a specific provider instead of
	// going through the task router — used by the e-mail analysis, where the
	// operator chooses the provider that may see private mail. Repopulated by
	// SwapNamed on every hot-swap, so a provider toggle is reflected live.
	named map[string]Provider
}

// NewManager creates a Manager wrapping an initial service (which may be nil).
func NewManager(initial ServiceInterface) *Manager {
	return &Manager{inner: initial}
}

// Swap atomically replaces the underlying service.
func (m *Manager) Swap(svc ServiceInterface) {
	m.mu.Lock()
	m.inner = svc
	m.mu.Unlock()
}

// SwapNamed atomically replaces the by-name provider registry.
func (m *Manager) SwapNamed(named map[string]Provider) {
	m.mu.Lock()
	m.named = named
	m.mu.Unlock()
}

// Named returns the enabled provider registered under name, or nil when no such
// provider is currently enabled. Callers that pin a provider treat nil as
// "AI unavailable" rather than silently falling back to another provider.
func (m *Manager) Named(name string) Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.named == nil {
		return nil
	}
	return m.named[name]
}

// Enabled reports whether a backing service is currently configured.
func (m *Manager) Enabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.inner != nil
}

func (m *Manager) Execute(ctx context.Context, req Request) (Response, error) {
	m.mu.RLock()
	inner := m.inner
	m.mu.RUnlock()

	if inner == nil {
		return Response{}, ErrServiceUnavailable
	}
	return inner.Execute(ctx, req)
}

// ExecuteStream forwards to the inner service's streaming capability when it has
// one; otherwise it falls back to Execute and emits the whole answer as a single
// chunk, keeping a uniform streaming contract for callers.
func (m *Manager) ExecuteStream(ctx context.Context, req Request, onChunk StreamFunc) (Response, error) {
	m.mu.RLock()
	inner := m.inner
	m.mu.RUnlock()

	if inner == nil {
		return Response{}, ErrServiceUnavailable
	}

	if streamer, ok := inner.(StreamingServiceInterface); ok {
		return streamer.ExecuteStream(ctx, req, onChunk)
	}

	resp, err := inner.Execute(ctx, req)
	if err != nil {
		return Response{}, err
	}
	if resp.Content != "" {
		if cbErr := onChunk(resp.Content); cbErr != nil {
			return Response{}, cbErr
		}
	}
	return resp, nil
}
