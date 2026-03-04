INSERT INTO video_playlist (
    type,
    source_path
) VALUES (
    $1,
    $2
)
RETURNING id, created_at, updated_at, last_played_at;
