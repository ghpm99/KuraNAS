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
