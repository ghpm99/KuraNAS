INSERT INTO playlist (name, description, is_system, kind, rules)
VALUES ($1, $2, FALSE, 'rule', $3::jsonb)
RETURNING id, created_at, updated_at;
