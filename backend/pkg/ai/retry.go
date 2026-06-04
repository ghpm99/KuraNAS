package ai

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// retryProvider decorates a Provider with per-provider retry behaviour. Retry
// configuration is owned by each provider (sourced from the ai_providers
// table) rather than being a single global policy, so a slow local provider
// and a fast cloud provider can be tuned independently.
type retryProvider struct {
	inner      Provider
	maxRetries int
	backoff    time.Duration
}

// WithRetry wraps a provider so failed, retryable calls are retried up to
// maxRetries times with a linear backoff. When maxRetries <= 0 the provider is
// returned unwrapped.
func WithRetry(inner Provider, maxRetries int, backoff time.Duration) Provider {
	if maxRetries <= 0 {
		return inner
	}
	return &retryProvider{inner: inner, maxRetries: maxRetries, backoff: backoff}
}

func (p *retryProvider) Name() string { return p.inner.Name() }

func (p *retryProvider) Complete(ctx context.Context, req Request) (Response, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Response{}, fmt.Errorf("%w: %v", ErrProviderTimeout, ctx.Err())
			case <-time.After(p.backoff * time.Duration(attempt)):
			}
		}

		resp, err := p.inner.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryable(err) {
			return Response{}, err
		}
	}

	return Response{}, fmt.Errorf("%w: %v", ErrAllAttemptsFailed, lastErr)
}

// CompleteStream forwards streaming to the wrapped provider when it supports it.
// Streaming is attempted once (no mid-stream retry), since partial output cannot
// be safely replayed; non-streaming calls keep their full retry behaviour.
func (p *retryProvider) CompleteStream(ctx context.Context, req Request, onChunk StreamFunc) (Response, error) {
	if streamer, ok := p.inner.(StreamingProvider); ok {
		return streamer.CompleteStream(ctx, req, onChunk)
	}
	return Response{}, ErrStreamingUnsupported
}

func isRetryable(err error) bool {
	return errors.Is(err, ErrProviderTimeout) || errors.Is(err, ErrProviderRateLimit)
}
