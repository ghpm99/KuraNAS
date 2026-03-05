SELECT
    id,
    type,
    source_path,
    name,
    is_hidden,
    is_auto,
    group_mode,
    classification,
    created_at,
    updated_at,
    last_played_at
FROM video_playlist
WHERE id = $1
LIMIT 1;
