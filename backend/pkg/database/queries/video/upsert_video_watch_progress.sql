INSERT INTO video_watch_progress (
    client_id,
    video_id,
    position_seconds,
    duration_seconds,
    completed,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    NOW()
)
ON CONFLICT (client_id, video_id)
DO UPDATE SET
    position_seconds = EXCLUDED.position_seconds,
    duration_seconds = EXCLUDED.duration_seconds,
    completed = EXCLUDED.completed,
    updated_at = NOW()
RETURNING updated_at;
