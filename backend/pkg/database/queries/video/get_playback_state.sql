SELECT
    id,
    client_id,
    playlist_id,
    video_id,
    current_position,
    duration,
    is_paused,
    completed,
    last_update
FROM video_playback_state
WHERE client_id = $1
LIMIT 1;
