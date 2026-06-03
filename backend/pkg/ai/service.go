package ai

import (
	"context"
	"fmt"
)

// Service orchestrates AI requests by resolving the appropriate provider chain
// via the Router and walking it (primary then fallbacks). Retry/timeout policy
// is owned per provider (see WithRetry), not by the Service.
type Service struct {
	router *Router
}

// NewService creates an AI service for the given router.
func NewService(router *Router) ServiceInterface {
	return &Service{router: router}
}

func (s *Service) Execute(ctx context.Context, req Request) (Response, error) {
	if req.Prompt == "" {
		return Response{}, ErrEmptyPrompt
	}

	route, err := s.router.Resolve(req.TaskType)
	if err != nil {
		return Response{}, err
	}

	resp, err := route.Primary.Complete(ctx, req)
	if err == nil {
		return resp, nil
	}

	lastErr := err
	for _, fallback := range route.Fallbacks {
		resp, err = fallback.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = fmt.Errorf("fallback provider %s failed: %w", fallback.Name(), err)
	}

	return Response{}, lastErr
}
