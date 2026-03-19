package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestConfig() Config {
	return Config{
		MaxRetries:     1,
		RetryBackoffMS: 10,
	}
}

func TestServiceExecuteSuccess(t *testing.T) {
	provider := &providerMock{
		name: "test",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{
				Content:  "generated text",
				Model:    "test-model",
				Provider: "test",
			}, nil
		},
	}

	router := NewRouter()
	router.Register(TaskGeneration, provider)
	service := NewService(router, newTestConfig())

	resp, err := service.Execute(context.Background(), Request{
		TaskType: TaskGeneration,
		Prompt:   "test prompt",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "generated text" {
		t.Fatalf("expected 'generated text', got %s", resp.Content)
	}
	if resp.Provider != "test" {
		t.Fatalf("expected provider 'test', got %s", resp.Provider)
	}
}

func TestServiceExecuteEmptyPrompt(t *testing.T) {
	router := NewRouter()
	service := NewService(router, newTestConfig())

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskGeneration,
		Prompt:   "",
	})
	if !errors.Is(err, ErrEmptyPrompt) {
		t.Fatalf("expected ErrEmptyPrompt, got %v", err)
	}
}

func TestServiceExecuteNoProvider(t *testing.T) {
	router := NewRouter()
	service := NewService(router, newTestConfig())

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskGeneration,
		Prompt:   "test",
	})
	if !errors.Is(err, ErrNoProviderForTask) {
		t.Fatalf("expected ErrNoProviderForTask, got %v", err)
	}
}

func TestServiceExecuteRetryOnTimeout(t *testing.T) {
	callCount := 0
	provider := &providerMock{
		name: "retry-test",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			callCount++
			if callCount < 2 {
				return Response{}, ErrProviderTimeout
			}
			return Response{Content: "success", Provider: "retry-test"}, nil
		},
	}

	router := NewRouter()
	router.Register(TaskSimple, provider)
	service := NewService(router, newTestConfig())

	resp, err := service.Execute(context.Background(), Request{
		TaskType: TaskSimple,
		Prompt:   "retry me",
	})
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if resp.Content != "success" {
		t.Fatalf("expected 'success', got %s", resp.Content)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls, got %d", callCount)
	}
}

func TestServiceExecuteRetryOnRateLimit(t *testing.T) {
	callCount := 0
	provider := &providerMock{
		name: "rate-limit-test",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			callCount++
			if callCount < 2 {
				return Response{}, ErrProviderRateLimit
			}
			return Response{Content: "ok", Provider: "rate-limit-test"}, nil
		},
	}

	router := NewRouter()
	router.Register(TaskSimple, provider)
	service := NewService(router, newTestConfig())

	resp, err := service.Execute(context.Background(), Request{
		TaskType: TaskSimple,
		Prompt:   "test",
	})
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("expected 'ok', got %s", resp.Content)
	}
}

func TestServiceExecuteNoRetryOnAuthError(t *testing.T) {
	callCount := 0
	provider := &providerMock{
		name: "auth-fail",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			callCount++
			return Response{}, ErrProviderAuth
		},
	}

	router := NewRouter()
	router.Register(TaskSimple, provider)
	service := NewService(router, newTestConfig())

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskSimple,
		Prompt:   "test",
	})
	if !errors.Is(err, ErrProviderAuth) {
		t.Fatalf("expected ErrProviderAuth, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call (no retry on auth error), got %d", callCount)
	}
}

func TestServiceExecuteAllRetriesFailed(t *testing.T) {
	provider := &providerMock{
		name: "always-timeout",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, ErrProviderTimeout
		},
	}

	router := NewRouter()
	router.Register(TaskSimple, provider)
	service := NewService(router, newTestConfig())

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskSimple,
		Prompt:   "test",
	})
	if !errors.Is(err, ErrAllAttemptsFailed) {
		t.Fatalf("expected ErrAllAttemptsFailed, got %v", err)
	}
}

func TestServiceExecuteFallback(t *testing.T) {
	primary := &providerMock{
		name: "primary",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, errors.New("primary down")
		},
	}
	fallback := &providerMock{
		name: "fallback",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{Content: "fallback response", Provider: "fallback"}, nil
		},
	}

	router := NewRouter()
	router.RegisterWithFallback(TaskComplex, primary, fallback)
	service := NewService(router, newTestConfig())

	resp, err := service.Execute(context.Background(), Request{
		TaskType: TaskComplex,
		Prompt:   "test",
	})
	if err != nil {
		t.Fatalf("expected fallback success, got %v", err)
	}
	if resp.Provider != "fallback" {
		t.Fatalf("expected fallback provider, got %s", resp.Provider)
	}
}

func TestServiceExecuteFallbackAlsoFails(t *testing.T) {
	primary := &providerMock{
		name: "primary",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, errors.New("primary down")
		},
	}
	fallback := &providerMock{
		name: "fallback",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, errors.New("fallback also down")
		},
	}

	router := NewRouter()
	router.RegisterWithFallback(TaskComplex, primary, fallback)
	service := NewService(router, newTestConfig())

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskComplex,
		Prompt:   "test",
	})
	if err == nil {
		t.Fatalf("expected error when both providers fail")
	}
}

func TestServiceExecuteContextCancelled(t *testing.T) {
	provider := &providerMock{
		name: "slow",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, ErrProviderTimeout
		},
	}

	router := NewRouter()
	router.Register(TaskSimple, provider)
	cfg := newTestConfig()
	cfg.RetryBackoffMS = 500
	service := NewService(router, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := service.Execute(ctx, Request{
		TaskType: TaskSimple,
		Prompt:   "test",
	})
	if err == nil {
		t.Fatalf("expected error on cancelled context")
	}
}
