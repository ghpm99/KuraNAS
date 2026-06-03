package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetryNoWrapWhenDisabled(t *testing.T) {
	provider := &providerMock{name: "p"}
	if WithRetry(provider, 0, time.Millisecond) != Provider(provider) {
		t.Fatalf("expected provider returned unwrapped when maxRetries <= 0")
	}
}

func TestRetryProviderName(t *testing.T) {
	p := WithRetry(&providerMock{name: "ollama"}, 2, time.Millisecond)
	if p.Name() != "ollama" {
		t.Fatalf("expected delegated name 'ollama', got %s", p.Name())
	}
}

func TestRetryProviderRetriesUntilSuccess(t *testing.T) {
	calls := 0
	provider := &providerMock{
		name: "retry",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			calls++
			if calls < 2 {
				return Response{}, ErrProviderTimeout
			}
			return Response{Content: "ok", Provider: "retry"}, nil
		},
	}

	resp, err := WithRetry(provider, 2, time.Millisecond).Complete(context.Background(), Request{Prompt: "x"})
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if resp.Content != "ok" || calls != 2 {
		t.Fatalf("expected 2 calls and success, got calls=%d resp=%+v", calls, resp)
	}
}

func TestRetryProviderRetriesOnRateLimit(t *testing.T) {
	calls := 0
	provider := &providerMock{
		name: "retry",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			calls++
			if calls < 2 {
				return Response{}, ErrProviderRateLimit
			}
			return Response{Content: "ok"}, nil
		},
	}

	if _, err := WithRetry(provider, 2, time.Millisecond).Complete(context.Background(), Request{Prompt: "x"}); err != nil {
		t.Fatalf("expected success after rate-limit retry, got %v", err)
	}
}

func TestRetryProviderDoesNotRetryNonRetryable(t *testing.T) {
	calls := 0
	provider := &providerMock{
		name: "auth",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			calls++
			return Response{}, ErrProviderAuth
		},
	}

	_, err := WithRetry(provider, 3, time.Millisecond).Complete(context.Background(), Request{Prompt: "x"})
	if !errors.Is(err, ErrProviderAuth) {
		t.Fatalf("expected ErrProviderAuth, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected a single attempt for non-retryable error, got %d", calls)
	}
}

func TestRetryProviderExhaustsAttempts(t *testing.T) {
	provider := &providerMock{
		name: "always-timeout",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, ErrProviderTimeout
		},
	}

	_, err := WithRetry(provider, 1, time.Millisecond).Complete(context.Background(), Request{Prompt: "x"})
	if !errors.Is(err, ErrAllAttemptsFailed) {
		t.Fatalf("expected ErrAllAttemptsFailed, got %v", err)
	}
}

func TestRetryProviderContextCancelled(t *testing.T) {
	provider := &providerMock{
		name: "slow",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{}, ErrProviderTimeout
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := WithRetry(provider, 3, 500*time.Millisecond).Complete(ctx, Request{Prompt: "x"})
	if err == nil {
		t.Fatalf("expected error on cancelled context")
	}
}
