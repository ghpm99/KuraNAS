package ai

import (
	"context"
	"errors"
	"testing"
)

func TestManagerExecuteWithoutInner(t *testing.T) {
	m := NewManager(nil)
	if m.Enabled() {
		t.Fatalf("expected manager to be disabled with nil inner")
	}

	_, err := m.Execute(context.Background(), Request{TaskType: TaskSimple, Prompt: "hi"})
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("expected ErrServiceUnavailable, got %v", err)
	}
}

func TestManagerSwapAndExecute(t *testing.T) {
	router := NewRouter()
	router.Register(TaskSimple, &providerMock{name: "p"})
	svc := NewService(router)

	m := NewManager(nil)
	m.Swap(svc)

	if !m.Enabled() {
		t.Fatalf("expected manager to be enabled after swap")
	}

	resp, err := m.Execute(context.Background(), Request{TaskType: TaskSimple, Prompt: "hi"})
	if err != nil {
		t.Fatalf("expected no error after swap, got %v", err)
	}
	if resp.Provider != "p" {
		t.Fatalf("expected provider 'p', got %s", resp.Provider)
	}
}

func TestManagerSwapToNilDisables(t *testing.T) {
	router := NewRouter()
	router.Register(TaskSimple, &providerMock{name: "p"})
	m := NewManager(NewService(router))

	m.Swap(nil)
	if m.Enabled() {
		t.Fatalf("expected manager disabled after swapping nil")
	}
	if _, err := m.Execute(context.Background(), Request{TaskType: TaskSimple, Prompt: "x"}); !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("expected ErrServiceUnavailable after nil swap, got %v", err)
	}
}

func TestManagerNamedLookupAndHotSwap(t *testing.T) {
	m := NewManager(nil)

	// Nothing registered yet.
	if m.Named("ollama") != nil {
		t.Fatalf("expected nil before any named provider is registered")
	}

	ollama := &providerMock{name: "ollama"}
	m.SwapNamed(map[string]Provider{"ollama": ollama})

	if got := m.Named("ollama"); got != ollama {
		t.Fatalf("Named(ollama) = %v, want the registered provider", got)
	}
	if m.Named("anthropic") != nil {
		t.Fatalf("expected nil for an unregistered name")
	}

	// Hot-swap: enabling anthropic and dropping ollama must reflect live.
	anthropic := &providerMock{name: "anthropic"}
	m.SwapNamed(map[string]Provider{"anthropic": anthropic})

	if m.Named("ollama") != nil {
		t.Fatalf("ollama should be gone after the swap")
	}
	if got := m.Named("anthropic"); got != anthropic {
		t.Fatalf("Named(anthropic) = %v, want the swapped-in provider", got)
	}
}
