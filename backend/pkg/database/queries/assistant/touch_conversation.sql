UPDATE assistant_conversations
SET updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
