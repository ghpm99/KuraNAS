INSERT INTO assistant_conversations (title)
VALUES ($1)
RETURNING id, title, created_at, updated_at;
