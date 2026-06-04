INSERT INTO playlist (name, description, is_system, is_ai_generated)
VALUES ($1, $2, FALSE, TRUE)
RETURNING id, created_at, updated_at;
