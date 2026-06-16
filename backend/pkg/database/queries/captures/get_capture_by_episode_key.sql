SELECT id, name, file_name, file_path, media_type, mime_type, size, episode_key, created_at
FROM captures
WHERE episode_key = $1
ORDER BY created_at DESC
LIMIT 1;
