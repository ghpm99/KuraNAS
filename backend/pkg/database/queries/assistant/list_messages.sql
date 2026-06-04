SELECT id, conversation_id, role, content, created_at
FROM assistant_messages
WHERE conversation_id = $1
ORDER BY id ASC;
