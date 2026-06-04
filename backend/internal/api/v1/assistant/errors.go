package assistant

import "errors"

var (
	// ErrAIUnavailable means no AI provider is configured/enabled, so the
	// assistant cannot answer.
	ErrAIUnavailable = errors.New("assistant: no AI provider available")
	// ErrInvalidConversation means the message list is empty or does not end
	// with a user turn.
	ErrInvalidConversation = errors.New("assistant: invalid conversation")
	// ErrEmptyResponse means the model returned no usable content.
	ErrEmptyResponse = errors.New("assistant: empty response from model")
	// ErrConversationNotFound means the referenced conversation does not exist.
	ErrConversationNotFound = errors.New("assistant: conversation not found")
)
