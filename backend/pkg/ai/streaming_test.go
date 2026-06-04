package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// streamingProviderMock implements both Provider and StreamingProvider.
type streamingProviderMock struct {
	name       string
	deltas     []string
	resp       Response
	streamErr  error
	completeFn func(ctx context.Context, req Request) (Response, error)
}

func (m *streamingProviderMock) Name() string { return m.name }

func (m *streamingProviderMock) Complete(ctx context.Context, req Request) (Response, error) {
	if m.completeFn != nil {
		return m.completeFn(ctx, req)
	}
	return m.resp, nil
}

func (m *streamingProviderMock) CompleteStream(ctx context.Context, req Request, onChunk StreamFunc) (Response, error) {
	for _, delta := range m.deltas {
		if err := onChunk(delta); err != nil {
			return Response{}, err
		}
	}
	if m.streamErr != nil {
		return Response{}, m.streamErr
	}
	return m.resp, nil
}

// nonStreamingService implements only ServiceInterface (no ExecuteStream).
type nonStreamingService struct {
	resp Response
	err  error
}

func (s *nonStreamingService) Execute(ctx context.Context, req Request) (Response, error) {
	return s.resp, s.err
}

func collectStream(t *testing.T, fn func(StreamFunc) (Response, error)) (string, Response, error) {
	t.Helper()
	var b strings.Builder
	resp, err := fn(func(chunk string) error {
		b.WriteString(chunk)
		return nil
	})
	return b.String(), resp, err
}

func TestServiceExecuteStreamEmptyPrompt(t *testing.T) {
	service := NewService(NewRouter()).(*Service)

	_, err := service.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration}, func(string) error { return nil })

	if !errors.Is(err, ErrEmptyPrompt) {
		t.Fatalf("expected ErrEmptyPrompt, got %v", err)
	}
}

func TestServiceExecuteStreamNoProvider(t *testing.T) {
	service := NewService(NewRouter()).(*Service)

	_, err := service.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, func(string) error { return nil })

	if !errors.Is(err, ErrNoProviderForTask) {
		t.Fatalf("expected ErrNoProviderForTask, got %v", err)
	}
}

func TestServiceExecuteStreamWithStreamingProvider(t *testing.T) {
	provider := &streamingProviderMock{
		name:   "stream",
		deltas: []string{"he", "llo"},
		resp:   Response{Content: "hello", Provider: "stream"},
	}
	router := NewRouter()
	router.Register(TaskGeneration, provider)
	service := NewService(router).(*Service)

	text, resp, err := collectStream(t, func(cb StreamFunc) (Response, error) {
		return service.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, cb)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "hello" || resp.Content != "hello" {
		t.Fatalf("expected reassembled 'hello', got text=%q resp=%q", text, resp.Content)
	}
}

func TestServiceExecuteStreamFallsBackWhenUnsupported(t *testing.T) {
	// providerMock implements Provider but not StreamingProvider.
	provider := &providerMock{
		name: "plain",
		completeFn: func(ctx context.Context, req Request) (Response, error) {
			return Response{Content: "full answer", Provider: "plain"}, nil
		},
	}
	router := NewRouter()
	router.Register(TaskGeneration, provider)
	service := NewService(router).(*Service)

	text, _, err := collectStream(t, func(cb StreamFunc) (Response, error) {
		return service.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, cb)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "full answer" {
		t.Fatalf("expected single full-content chunk, got %q", text)
	}
}

func TestServiceExecuteStreamPropagatesStreamError(t *testing.T) {
	provider := &streamingProviderMock{name: "stream", streamErr: errors.New("boom")}
	router := NewRouter()
	router.Register(TaskGeneration, provider)
	service := NewService(router).(*Service)

	_, err := service.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected propagated stream error, got %v", err)
	}
}

func TestManagerExecuteStreamNilInner(t *testing.T) {
	manager := NewManager(nil)

	_, err := manager.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, func(string) error { return nil })
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("expected ErrServiceUnavailable, got %v", err)
	}
}

func TestManagerExecuteStreamForwardsToStreamingInner(t *testing.T) {
	provider := &streamingProviderMock{name: "stream", deltas: []string{"a", "b"}, resp: Response{Content: "ab"}}
	router := NewRouter()
	router.Register(TaskGeneration, provider)
	manager := NewManager(NewService(router))

	text, _, err := collectStream(t, func(cb StreamFunc) (Response, error) {
		return manager.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, cb)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "ab" {
		t.Fatalf("expected 'ab', got %q", text)
	}
}

func TestManagerExecuteStreamFallsBackForNonStreamingInner(t *testing.T) {
	manager := NewManager(&nonStreamingService{resp: Response{Content: "whole"}})

	text, _, err := collectStream(t, func(cb StreamFunc) (Response, error) {
		return manager.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, cb)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "whole" {
		t.Fatalf("expected single 'whole' chunk, got %q", text)
	}
}

func TestManagerExecuteStreamNonStreamingInnerError(t *testing.T) {
	manager := NewManager(&nonStreamingService{err: errors.New("down")})

	_, err := manager.ExecuteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "down") {
		t.Fatalf("expected inner error, got %v", err)
	}
}

func TestRetryProviderCompleteStreamForwards(t *testing.T) {
	inner := &streamingProviderMock{name: "stream", deltas: []string{"x", "y"}, resp: Response{Content: "xy"}}
	wrapped := WithRetry(inner, 2, 0)

	streamer, ok := wrapped.(StreamingProvider)
	if !ok {
		t.Fatal("retry provider should expose streaming")
	}
	text, _, err := collectStream(t, func(cb StreamFunc) (Response, error) {
		return streamer.CompleteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, cb)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "xy" {
		t.Fatalf("expected 'xy', got %q", text)
	}
}

func TestRetryProviderCompleteStreamUnsupportedInner(t *testing.T) {
	inner := &providerMock{name: "plain", completeFn: func(ctx context.Context, req Request) (Response, error) {
		return Response{}, nil
	}}
	wrapped := WithRetry(inner, 2, 0).(StreamingProvider)

	_, err := wrapped.CompleteStream(context.Background(), Request{TaskType: TaskGeneration, Prompt: "x"}, func(string) error { return nil })
	if !errors.Is(err, ErrStreamingUnsupported) {
		t.Fatalf("expected ErrStreamingUnsupported, got %v", err)
	}
}
