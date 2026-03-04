INSERT INTO video_playlist (
    type,
    source_path,
    name,
    is_hidden,
    is_auto,
    group_mode,
    classification,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    FALSE,
    TRUE,
    $4,
    $5,
    NOW(),
    NOW()
)
ON CONFLICT (type, source_path)
DO UPDATE SET
    name = EXCLUDED.name,
    is_auto = TRUE,
    group_mode = EXCLUDED.group_mode,
    classification = EXCLUDED.classification,
    updated_at = NOW()
RETURNING id, type, source_path, name, is_hidden, is_auto, group_mode, classification, created_at, updated_at, last_played_at;
