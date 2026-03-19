package ai

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Service orchestrates AI requests by resolving the appropriate provider
// via the Router and handling retry/fallback logic.
type Service struct {
	router     *Router
	maxRetries int
	backoff    time.Duration
}

// NewService creates an AI service with the given router and config.
func NewService(router *Router, cfg Config) ServiceInterface {
	return &Service{
		router:     router,
		maxRetries: cfg.MaxRetries,
		backoff:    time.Duration(cfg.RetryBackoffMS) * time.Millisecond,
	}
}

func (s *Service) Execute(ctx context.Context, req Request) (Response, error) {
	if req.Prompt == "" {
		return Response{}, ErrEmptyPrompt
	}

	route, err := s.router.Resolve(req.TaskType)
	if err != nil {
		return Response{}, err
	}

	resp, err := s.executeWithRetry(ctx, route.Primary, req)
	if err != nil && route.Fallback != nil {
		resp, err = s.executeWithRetry(ctx, route.Fallback, req)
		if err != nil {
			return Response{}, fmt.Errorf("fallback provider %s failed: %w", route.Fallback.Name(), err)
		}
	}

	return resp, err
}

func (s *Service) executeWithRetry(ctx context.Context, provider Provider, req Request) (Response, error) {
	var lastErr error

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Response{}, fmt.Errorf("%w: %v", ErrProviderTimeout, ctx.Err())
			case <-time.After(s.backoff * time.Duration(attempt)):
			}
		}

		resp, err := provider.Complete(ctx, req)
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

func isRetryable(err error) bool {
	return errors.Is(err, ErrProviderTimeout) || errors.Is(err, ErrProviderRateLimit)
}
