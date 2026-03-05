UPDATE video_playlist
SET is_hidden = $2,
    updated_at = NOW()
WHERE id = $1;
