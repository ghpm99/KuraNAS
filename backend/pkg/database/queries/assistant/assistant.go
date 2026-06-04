package queries

import (
	_ "embed"
)

//go:embed insert_conversation.sql
var InsertConversationQuery string

//go:embed list_conversations.sql
var ListConversationsQuery string

//go:embed conversation_exists.sql
var ConversationExistsQuery string

//go:embed touch_conversation.sql
var TouchConversationQuery string

//go:embed delete_conversation.sql
var DeleteConversationQuery string

//go:embed insert_message.sql
var InsertMessageQuery string

//go:embed list_messages.sql
var ListMessagesQuery string
