SELECT
    file_id,
    season,
    episode
FROM video_manual_override
WHERE file_id = ANY($1::int[])
  AND (season IS NOT NULL OR episode IS NOT NULL);
