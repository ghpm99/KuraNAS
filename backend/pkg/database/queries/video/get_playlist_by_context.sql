SELECT
    id,
    type,
    source_path,
    created_at,
    updated_at,
    last_played_at
FROM video_playlist
WHERE type = $1
  AND source_path = $2
LIMIT 1;
