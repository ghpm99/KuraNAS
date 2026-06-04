package assistant

// Role identifies who authored a chat message.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ChatMessageDto is a single turn returned in a chat reply.
type ChatMessageDto struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequestDto is the body of the chat endpoints. The client sends the new
// user message and, on a follow-up turn, the conversation it belongs to; the
// backend loads the prior history from storage.
type ChatRequestDto struct {
	ConversationID int    `json:"conversation_id,omitempty"`
	Message        string `json:"message"`
}

// ChatResponseDto is the assistant's reply plus the conversation it was stored
// in and traceability of which model/provider produced it.
type ChatResponseDto struct {
	ConversationID int            `json:"conversation_id"`
	Message        ChatMessageDto `json:"message"`
	Model          string         `json:"model"`
	Provider       string         `json:"provider"`
}

// ConversationDto is a stored conversation in the transport shape.
type ConversationDto struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MessageDto is a stored message in the transport shape.
type MessageDto struct {
	ID        int    `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// StreamDeltaDto is the payload of a streaming `delta` event.
type StreamDeltaDto struct {
	Content string `json:"content"`
}

// StreamErrorDto is the payload of a streaming `error` event.
type StreamErrorDto struct {
	Error string `json:"error"`
}
