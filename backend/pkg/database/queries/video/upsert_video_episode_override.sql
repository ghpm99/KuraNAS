INSERT INTO video_manual_override (file_id, season, episode)
VALUES ($1, $2, $3)
ON CONFLICT (file_id) DO UPDATE
SET season = EXCLUDED.season,
    episode = EXCLUDED.episode,
    updated_at = now();
