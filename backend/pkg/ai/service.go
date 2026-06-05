package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nas-go/api/pkg/applog"
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

	started := time.Now()
	resp, err := route.Primary.Complete(ctx, req)
	if err == nil {
		applog.Debug("ai request completed",
			"task", string(req.TaskType), "provider", route.Primary.Name(),
			"latency_ms", time.Since(started).Milliseconds())
		return resp, nil
	}
	applog.Warn("ai provider failed, trying fallbacks",
		"task", string(req.TaskType), "provider", route.Primary.Name(),
		"latency_ms", time.Since(started).Milliseconds(), "error", err.Error())

	lastErr := err
	for _, fallback := range route.Fallbacks {
		fbStart := time.Now()
		resp, err = fallback.Complete(ctx, req)
		if err == nil {
			applog.Debug("ai request completed via fallback",
				"task", string(req.TaskType), "provider", fallback.Name(),
				"latency_ms", time.Since(fbStart).Milliseconds())
			return resp, nil
		}
		applog.Warn("ai fallback provider failed",
			"task", string(req.TaskType), "provider", fallback.Name(), "error", err.Error())
		lastErr = fmt.Errorf("fallback provider %s failed: %w", fallback.Name(), err)
	}

	applog.Error("ai request failed on all providers",
		"task", string(req.TaskType), "error", lastErr.Error())
	return Response{}, lastErr
}

// ExecuteStream streams the primary provider's output through onChunk. When the
// provider cannot stream (or streaming fails before producing output), it falls
// back to a non-streaming Execute and delivers the whole answer as one chunk, so
// callers get a uniform streaming contract regardless of provider capability.
func (s *Service) ExecuteStream(ctx context.Context, req Request, onChunk StreamFunc) (Response, error) {
	if req.Prompt == "" {
		return Response{}, ErrEmptyPrompt
	}

	route, err := s.router.Resolve(req.TaskType)
	if err != nil {
		return Response{}, err
	}

	if streamer, ok := route.Primary.(StreamingProvider); ok {
		resp, streamErr := streamer.CompleteStream(ctx, req, onChunk)
		if streamErr == nil {
			return resp, nil
		}
		if !errors.Is(streamErr, ErrStreamingUnsupported) {
			return Response{}, streamErr
		}
	}

	resp, err := s.Execute(ctx, req)
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
