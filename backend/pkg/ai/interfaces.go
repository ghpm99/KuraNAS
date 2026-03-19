package ai

import "context"

// Provider abstracts a single AI model/provider integration.
// Each implementation encapsulates the HTTP communication,
// request/response mapping, and authentication for one provider.
type Provider interface {
	Complete(ctx context.Context, req Request) (Response, error)
	Name() string
}

// ServiceInterface is the main entry point for AI operations.
// Other services in the application depend on this interface,
// never on concrete providers.
type ServiceInterface interface {
	Execute(ctx context.Context, req Request) (Response, error)
}
