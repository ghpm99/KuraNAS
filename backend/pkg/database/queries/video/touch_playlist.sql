UPDATE video_playlist
SET
    updated_at = NOW(),
    last_played_at = NOW()
WHERE id = $1;
