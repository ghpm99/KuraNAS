UPDATE playlist
SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3
RETURNING updated_at;
