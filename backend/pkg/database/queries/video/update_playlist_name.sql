UPDATE video_playlist
SET name = $2,
    name_key = NULL,
    updated_at = NOW()
WHERE id = $1;
