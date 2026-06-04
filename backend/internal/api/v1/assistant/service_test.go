package assistant

import (
	"context"
	"errors"
	"nas-go/api/pkg/ai"
	"strings"
	"testing"
)

type fakeAIService struct {
	resp   ai.Response
	err    error
	called bool
	gotReq ai.Request
}

func (f *fakeAIService) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	f.called = true
	f.gotReq = req
	return f.resp, f.err
}

func TestChatNilAIService(t *testing.T) {
	service := NewService(nil)

	_, err := service.Chat([]ChatMessageDto{{Role: RoleUser, Content: "oi"}})

	if !errors.Is(err, ErrAIUnavailable) {
		t.Fatalf("expected ErrAIUnavailable, got %v", err)
	}
}

func TestChatEmptyConversation(t *testing.T) {
	service := NewService(&fakeAIService{})

	_, err := service.Chat(nil)

	if !errors.Is(err, ErrInvalidConversation) {
		t.Fatalf("expected ErrInvalidConversation, got %v", err)
	}
}

func TestChatOnlyBlankMessages(t *testing.T) {
	service := NewService(&fakeAIService{})

	_, err := service.Chat([]ChatMessageDto{{Role: RoleUser, Content: "   "}})

	if !errors.Is(err, ErrInvalidConversation) {
		t.Fatalf("expected ErrInvalidConversation, got %v", err)
	}
}

func TestChatLastMessageNotUser(t *testing.T) {
	service := NewService(&fakeAIService{})

	_, err := service.Chat([]ChatMessageDto{
		{Role: RoleUser, Content: "oi"},
		{Role: RoleAssistant, Content: "olá"},
	})

	if !errors.Is(err, ErrInvalidConversation) {
		t.Fatalf("expected ErrInvalidConversation, got %v", err)
	}
}

func TestChatSuccess(t *testing.T) {
	fake := &fakeAIService{resp: ai.Response{Content: "  Olá, tudo bem?  ", Model: "llama3.1", Provider: "ollama"}}
	service := NewService(fake)

	resp, err := service.Chat([]ChatMessageDto{
		{Role: RoleUser, Content: "oi"},
		{Role: RoleAssistant, Content: "olá"},
		{Role: RoleUser, Content: "tudo bem?"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message.Role != RoleAssistant {
		t.Fatalf("expected assistant role, got %q", resp.Message.Role)
	}
	if resp.Message.Content != "Olá, tudo bem?" {
		t.Fatalf("expected trimmed content, got %q", resp.Message.Content)
	}
	if resp.Model != "llama3.1" || resp.Provider != "ollama" {
		t.Fatalf("missing traceability: %+v", resp)
	}
	if !fake.called {
		t.Fatal("expected AI service to be called")
	}
	if fake.gotReq.TaskType != ai.TaskGeneration {
		t.Fatalf("expected TaskGeneration, got %v", fake.gotReq.TaskType)
	}
	if fake.gotReq.SystemPrompt == "" {
		t.Fatal("expected a system prompt")
	}
	if !strings.HasSuffix(fake.gotReq.Prompt, "Assistente:") {
		t.Fatalf("prompt should end inviting the assistant to continue, got %q", fake.gotReq.Prompt)
	}
	if !strings.Contains(fake.gotReq.Prompt, "Usuário: oi") || !strings.Contains(fake.gotReq.Prompt, "Assistente: olá") {
		t.Fatalf("prompt should contain the history transcript, got %q", fake.gotReq.Prompt)
	}
}

func TestChatAIError(t *testing.T) {
	fake := &fakeAIService{err: errors.New("boom")}
	service := NewService(fake)

	_, err := service.Chat([]ChatMessageDto{{Role: RoleUser, Content: "oi"}})

	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected propagated AI error, got %v", err)
	}
}

func TestChatEmptyAIResponse(t *testing.T) {
	fake := &fakeAIService{resp: ai.Response{Content: "   "}}
	service := NewService(fake)

	_, err := service.Chat([]ChatMessageDto{{Role: RoleUser, Content: "oi"}})

	if !errors.Is(err, ErrEmptyResponse) {
		t.Fatalf("expected ErrEmptyResponse, got %v", err)
	}
}

func TestChatDropsUnknownRoles(t *testing.T) {
	fake := &fakeAIService{resp: ai.Response{Content: "ok"}}
	service := NewService(fake)

	_, err := service.Chat([]ChatMessageDto{
		{Role: "system", Content: "ignore me"},
		{Role: RoleUser, Content: "oi"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(fake.gotReq.Prompt, "ignore me") {
		t.Fatalf("unknown-role message should have been dropped, got %q", fake.gotReq.Prompt)
	}
}

func TestChatCapsHistory(t *testing.T) {
	fake := &fakeAIService{resp: ai.Response{Content: "ok"}}
	service := NewService(fake)

	messages := make([]ChatMessageDto, 0, maxHistoryMessages+5)
	messages = append(messages, ChatMessageDto{Role: RoleUser, Content: "primeira-mensagem-antiga"})
	for i := 0; i < maxHistoryMessages+4; i++ {
		messages = append(messages, ChatMessageDto{Role: RoleUser, Content: "recente"})
	}

	if _, err := service.Chat(messages); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(fake.gotReq.Prompt, "primeira-mensagem-antiga") {
		t.Fatalf("oldest message should have been dropped by the cap, got %q", fake.gotReq.Prompt)
	}
}
