SELECT
    client_id,
    video_id,
    position_seconds,
    duration_seconds,
    completed,
    updated_at
FROM video_watch_progress
WHERE client_id = $1
  AND video_id = ANY($2);
