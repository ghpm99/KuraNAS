INSERT INTO playlist (name, description, is_system)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at;
