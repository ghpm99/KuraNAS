INSERT INTO playlist_rule_exclusion (playlist_id, file_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
