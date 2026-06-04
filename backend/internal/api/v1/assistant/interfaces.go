package assistant

// ServiceInterface is the chat entry point. This first iteration is
// conversation-only: no tools, no persistence.
type ServiceInterface interface {
	Chat(messages []ChatMessageDto) (ChatResponseDto, error)
}
