INSERT INTO playlist (name, description, is_system, is_ai_generated, kind)
VALUES ($1, $2, FALSE, TRUE, 'ai')
RETURNING id, created_at, updated_at;
