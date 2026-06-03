UPDATE ai_providers
SET enabled = $2,
    model = $3,
    base_url = $4,
    priority = $5,
    params = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE name = $1
RETURNING id, name, enabled, model, base_url, priority, params, created_at, updated_at;
