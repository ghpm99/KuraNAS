package ai

import (
	"context"
	"errors"
	"testing"
)

type providerMock struct {
	name       string
	completeFn func(ctx context.Context, req Request) (Response, error)
}

func (m *providerMock) Name() string { return m.name }
func (m *providerMock) Complete(ctx context.Context, req Request) (Response, error) {
	if m.completeFn != nil {
		return m.completeFn(ctx, req)
	}
	return Response{Content: "mock", Provider: m.name}, nil
}

func TestRouterRegisterAndResolve(t *testing.T) {
	router := NewRouter()
	provider := &providerMock{name: "test-provider"}

	router.Register(TaskGeneration, provider)

	route, err := router.Resolve(TaskGeneration)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if route.Primary.Name() != "test-provider" {
		t.Fatalf("expected test-provider, got %s", route.Primary.Name())
	}
	if len(route.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %d", len(route.Fallbacks))
	}
}

func TestRouterRegisterWithFallback(t *testing.T) {
	router := NewRouter()
	primary := &providerMock{name: "primary"}
	fallback := &providerMock{name: "fallback"}

	router.RegisterWithFallback(TaskComplex, primary, fallback)

	route, err := router.Resolve(TaskComplex)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if route.Primary.Name() != "primary" {
		t.Fatalf("expected primary, got %s", route.Primary.Name())
	}
	if len(route.Fallbacks) != 1 || route.Fallbacks[0].Name() != "fallback" {
		t.Fatalf("expected single fallback provider")
	}
}

func TestRouterRegisterChain(t *testing.T) {
	router := NewRouter()
	primary := &providerMock{name: "primary"}
	fb1 := &providerMock{name: "fb1"}
	fb2 := &providerMock{name: "fb2"}

	router.RegisterChain(TaskComplex, primary, fb1, fb2)

	route, err := router.Resolve(TaskComplex)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if route.Primary.Name() != "primary" {
		t.Fatalf("expected primary, got %s", route.Primary.Name())
	}
	if len(route.Fallbacks) != 2 {
		t.Fatalf("expected 2 fallbacks, got %d", len(route.Fallbacks))
	}
	if route.Fallbacks[0].Name() != "fb1" || route.Fallbacks[1].Name() != "fb2" {
		t.Fatalf("expected ordered fallbacks fb1, fb2, got %s, %s", route.Fallbacks[0].Name(), route.Fallbacks[1].Name())
	}
}

func TestRouterRegisterChainSingleProvider(t *testing.T) {
	router := NewRouter()
	router.RegisterChain(TaskSimple, &providerMock{name: "only"})

	route, err := router.Resolve(TaskSimple)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if route.Primary.Name() != "only" {
		t.Fatalf("expected only, got %s", route.Primary.Name())
	}
	if len(route.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %d", len(route.Fallbacks))
	}
}

func TestRouterRegisterChainEmpty(t *testing.T) {
	router := NewRouter()
	router.RegisterChain(TaskSimple)

	if _, err := router.Resolve(TaskSimple); !errors.Is(err, ErrNoProviderForTask) {
		t.Fatalf("expected ErrNoProviderForTask for empty chain, got %v", err)
	}
}

func TestRouterResolveUnregisteredTask(t *testing.T) {
	router := NewRouter()

	_, err := router.Resolve(TaskClassification)
	if err == nil {
		t.Fatalf("expected error for unregistered task")
	}
	if !errors.Is(err, ErrNoProviderForTask) {
		t.Fatalf("expected ErrNoProviderForTask, got %v", err)
	}
}

func TestRouterRegisteredTaskTypes(t *testing.T) {
	router := NewRouter()
	router.Register(TaskGeneration, &providerMock{name: "a"})
	router.Register(TaskSimple, &providerMock{name: "b"})

	types := router.RegisteredTaskTypes()
	if len(types) != 2 {
		t.Fatalf("expected 2 registered types, got %d", len(types))
	}

	found := map[TaskType]bool{}
	for _, tt := range types {
		found[tt] = true
	}
	if !found[TaskGeneration] || !found[TaskSimple] {
		t.Fatalf("expected generation and simple task types, got %v", types)
	}
}
