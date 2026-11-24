INSERT INTO playlist (name, description, is_system, kind)
VALUES ($1, $2, $3::boolean, CASE WHEN $3::boolean THEN 'system' ELSE 'manual' END)
RETURNING id, created_at, updated_at;
