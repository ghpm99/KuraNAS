INSERT INTO video_playlist_exclusion (
    playlist_id,
    video_id,
    created_at
)
VALUES ($1, $2, NOW())
ON CONFLICT (playlist_id, video_id)
DO NOTHING;
