DELETE FROM video_manual_override
WHERE file_id = $1
  AND classification IS NULL
  AND season IS NULL
  AND episode IS NULL;
