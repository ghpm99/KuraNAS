SELECT id, name, enabled, model, base_url, priority, params, created_at, updated_at
FROM ai_providers
WHERE name = $1;
