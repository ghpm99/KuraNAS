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
	if route.Fallback != nil {
		t.Fatalf("expected nil fallback")
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
	if route.Fallback == nil || route.Fallback.Name() != "fallback" {
		t.Fatalf("expected fallback provider")
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
