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
