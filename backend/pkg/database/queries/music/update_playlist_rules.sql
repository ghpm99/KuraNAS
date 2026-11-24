UPDATE playlist
SET rules = $1::jsonb, updated_at = CURRENT_TIMESTAMP
WHERE id = $2 AND kind = 'rule';
