INSERT INTO video_playback_state (
    client_id,
    playlist_id,
    video_id,
    current_position,
    duration,
    is_paused,
    completed,
    last_update
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    NOW()
)
ON CONFLICT (client_id)
DO UPDATE SET
    playlist_id = EXCLUDED.playlist_id,
    video_id = EXCLUDED.video_id,
    current_position = EXCLUDED.current_position,
    duration = EXCLUDED.duration,
    is_paused = EXCLUDED.is_paused,
    completed = EXCLUDED.completed,
    last_update = NOW()
RETURNING id, last_update;
