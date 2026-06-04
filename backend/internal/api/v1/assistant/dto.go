package assistant

// Role identifies who authored a chat message.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ChatMessageDto is a single turn in the conversation. The client holds the
// full history and sends it on every request; this first iteration does not
// persist conversations.
type ChatMessageDto struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequestDto is the body of POST /assistant/chat. Messages are ordered
// oldest-first and the last one must be from the user.
type ChatRequestDto struct {
	Messages []ChatMessageDto `json:"messages"`
}

// ChatResponseDto is the assistant's reply plus traceability of which
// model/provider produced it.
type ChatResponseDto struct {
	Message  ChatMessageDto `json:"message"`
	Model    string         `json:"model"`
	Provider string         `json:"provider"`
}
