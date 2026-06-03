package ai

import (
	"context"
	"errors"
	"testing"
)

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
	service := NewService(router)

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
	service := NewService(router)

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
	service := NewService(router)

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskGeneration,
		Prompt:   "test",
	})
	if !errors.Is(err, ErrNoProviderForTask) {
		t.Fatalf("expected ErrNoProviderForTask, got %v", err)
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
	service := NewService(router)

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
	service := NewService(router)

	_, err := service.Execute(context.Background(), Request{
		TaskType: TaskComplex,
		Prompt:   "test",
	})
	if err == nil {
		t.Fatalf("expected error when both providers fail")
	}
}

func TestServiceExecuteChainFallback(t *testing.T) {
	calls := []string{}
	makeProvider := func(name string, fail bool) *providerMock {
		return &providerMock{
			name: name,
			completeFn: func(ctx context.Context, req Request) (Response, error) {
				calls = append(calls, name)
				if fail {
					return Response{}, errors.New(name + " down")
				}
				return Response{Content: "ok", Provider: name}, nil
			},
		}
	}

	primary := makeProvider("primary", true)
	fb1 := makeProvider("fb1", true)
	fb2 := makeProvider("fb2", false)

	router := NewRouter()
	router.RegisterChain(TaskComplex, primary, fb1, fb2)
	service := NewService(router)

	resp, err := service.Execute(context.Background(), Request{TaskType: TaskComplex, Prompt: "test"})
	if err != nil {
		t.Fatalf("expected chain to succeed on fb2, got %v", err)
	}
	if resp.Provider != "fb2" {
		t.Fatalf("expected fb2 provider, got %s", resp.Provider)
	}
	if len(calls) != 3 {
		t.Fatalf("expected all 3 providers tried in order, got %v", calls)
	}
}

func TestServiceExecuteChainAllFail(t *testing.T) {
	makeProvider := func(name string) *providerMock {
		return &providerMock{
			name: name,
			completeFn: func(ctx context.Context, req Request) (Response, error) {
				return Response{}, errors.New(name + " down")
			},
		}
	}

	router := NewRouter()
	router.RegisterChain(TaskComplex, makeProvider("p"), makeProvider("fb1"), makeProvider("fb2"))
	service := NewService(router)

	_, err := service.Execute(context.Background(), Request{TaskType: TaskComplex, Prompt: "test"})
	if err == nil {
		t.Fatalf("expected error when every provider in the chain fails")
	}
}
