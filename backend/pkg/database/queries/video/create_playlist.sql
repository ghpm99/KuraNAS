INSERT INTO video_playlist (
    type,
    source_path,
    name,
    is_auto,
    group_mode,
    classification
) VALUES (
    $1,
    $2,
    $3,
    TRUE,
    'folder',
    'personal'
)
RETURNING id, type, source_path, name, is_hidden, is_auto, group_mode, classification, created_at, updated_at, last_played_at;
