INSERT INTO libraries (category, path, created_at, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (category)
DO UPDATE SET path = $2, updated_at = CURRENT_TIMESTAMP
RETURNING id, category, path, created_at, updated_at;
