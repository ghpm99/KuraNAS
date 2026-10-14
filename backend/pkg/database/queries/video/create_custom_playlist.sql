INSERT INTO video_playlist (
    type,
    source_path,
    name,
    is_auto,
    group_mode,
    classification
) VALUES (
    'custom',
    $1,
    $2,
    FALSE,
    'folder',
    'personal'
)
RETURNING id, type, source_path, name, is_hidden, is_auto, group_mode, classification, created_at, updated_at, last_played_at;
