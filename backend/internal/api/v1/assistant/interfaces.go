package assistant

import (
	"context"
	"nas-go/api/pkg/ai"
)

// DeltaFunc receives incremental content as the assistant produces it.
type DeltaFunc func(delta string) error

// AgentInterface is the optional tool-calling engine. When a message routes to a
// tool, the service delegates generation to the agent; otherwise it generates
// directly. Implemented by *agent.Agent.
type AgentInterface interface {
	HasToolsFor(message string) bool
	Generate(ctx context.Context, systemPrompt, prompt, message string, onDelta ai.StreamFunc) (ai.Response, error)
}

// ServiceInterface is the chat entry point. Conversation-only for now (no tools)
// but with persisted conversations and history.
type ServiceInterface interface {
	Chat(input ChatRequestDto) (ChatResponseDto, error)
	ChatStream(input ChatRequestDto, onDelta DeltaFunc) (ChatResponseDto, error)
	ListConversations() ([]ConversationDto, error)
	GetMessages(conversationID int) ([]MessageDto, error)
	DeleteConversation(conversationID int) error
}

// RepositoryInterface is the persistence boundary for conversations/messages.
type RepositoryInterface interface {
	CreateConversation(title string) (ConversationModel, error)
	ConversationExists(id int) (bool, error)
	ListConversations() ([]ConversationModel, error)
	TouchConversation(id int) error
	DeleteConversation(id int) error
	AddMessage(conversationID int, role, content string) (MessageModel, error)
	ListMessages(conversationID int) ([]MessageModel, error)
}
