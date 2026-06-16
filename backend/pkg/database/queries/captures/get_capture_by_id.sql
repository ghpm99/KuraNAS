SELECT id, name, file_name, file_path, media_type, mime_type, size, episode_key, created_at
FROM captures
WHERE id = $1;
