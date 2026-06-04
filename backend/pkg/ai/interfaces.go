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

// StreamFunc receives incremental content chunks as the model produces them.
// Returning an error aborts the generation.
type StreamFunc func(chunk string) error

// StreamingProvider is the optional capability a Provider exposes when it can
// emit content incrementally. Providers that do not implement it are driven via
// Complete and their output is delivered as a single chunk.
type StreamingProvider interface {
	CompleteStream(ctx context.Context, req Request, onChunk StreamFunc) (Response, error)
}

// StreamingServiceInterface is the optional streaming capability of a service.
// The Manager and the default Service both implement it; consumers type-assert
// to it and fall back to Execute when it is absent.
type StreamingServiceInterface interface {
	ServiceInterface
	ExecuteStream(ctx context.Context, req Request, onChunk StreamFunc) (Response, error)
}
