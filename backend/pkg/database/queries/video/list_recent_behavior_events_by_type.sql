SELECT
    id,
    client_id,
    video_id,
    COALESCE(playlist_id, 0) AS playlist_id,
    event_type,
    position,
    duration,
    watched_pct,
    created_at
FROM video_behavior_event
WHERE event_type = $1
ORDER BY created_at DESC
LIMIT $2;
