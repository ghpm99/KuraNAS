package assistant

import (
	"context"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"strings"
	"time"
)

const (
	// maxHistoryMessages caps how many turns are sent to the model, keeping the
	// prompt (and therefore the local-model cost) bounded on long conversations.
	maxHistoryMessages = 20
	// maxResponseTokens bounds the reply length.
	maxResponseTokens = 800
	// chatTemperature favours natural conversation over deterministic output.
	chatTemperature = 0.7
	// chatTimeout is the budget for a single reply from a local model.
	chatTimeout = 60 * time.Second
)

type Service struct {
	AIService ai.ServiceInterface
}

func NewService(aiService ai.ServiceInterface) ServiceInterface {
	return &Service{AIService: aiService}
}

func (s *Service) Chat(messages []ChatMessageDto) (ChatResponseDto, error) {
	if s.AIService == nil {
		return ChatResponseDto{}, ErrAIUnavailable
	}

	history := sanitizeMessages(messages)
	if len(history) == 0 || history[len(history)-1].Role != RoleUser {
		return ChatResponseDto{}, ErrInvalidConversation
	}

	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()

	resp, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskGeneration,
		SystemPrompt: prompts.AssistantChatSystemPrompt(),
		Prompt:       buildPrompt(history),
		MaxTokens:    maxResponseTokens,
		Temperature:  chatTemperature,
	})
	if err != nil {
		return ChatResponseDto{}, err
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return ChatResponseDto{}, ErrEmptyResponse
	}

	return ChatResponseDto{
		Message:  ChatMessageDto{Role: RoleAssistant, Content: content},
		Model:    resp.Model,
		Provider: resp.Provider,
	}, nil
}

// sanitizeMessages trims content, drops empty/unknown-role turns, and keeps only
// the most recent maxHistoryMessages so the prompt stays bounded.
func sanitizeMessages(messages []ChatMessageDto) []ChatMessageDto {
	cleaned := make([]ChatMessageDto, 0, len(messages))
	for _, m := range messages {
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		if m.Role != RoleUser && m.Role != RoleAssistant {
			continue
		}
		cleaned = append(cleaned, ChatMessageDto{Role: m.Role, Content: content})
	}

	if len(cleaned) > maxHistoryMessages {
		cleaned = cleaned[len(cleaned)-maxHistoryMessages:]
	}
	return cleaned
}

// buildPrompt flattens the conversation into a transcript the model continues as
// the assistant. The single-string shape fits the current ai.Request contract;
// multi-turn tool calling comes in a later iteration.
func buildPrompt(messages []ChatMessageDto) string {
	var b strings.Builder
	for _, m := range messages {
		label := "Usuário"
		if m.Role == RoleAssistant {
			label = "Assistente"
		}
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(m.Content)
		b.WriteString("\n")
	}
	b.WriteString("Assistente:")
	return b.String()
}
