package assistant

import "time"

// ConversationModel is the DB shape of a stored conversation.
type ConversationModel struct {
	ID        int
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MessageModel is the DB shape of a stored message.
type MessageModel struct {
	ID             int
	ConversationID int
	Role           string
	Content        string
	CreatedAt      time.Time
}
