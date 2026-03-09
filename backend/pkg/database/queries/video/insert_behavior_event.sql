INSERT INTO video_behavior_event (client_id, video_id, playlist_id, event_type, position, duration, watched_pct)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at;
