SELECT
    file.id,
    file.name,
    file.path,
    file.parent_path,
    file.format,
    file.size,
    file.created_at,
    file.updated_at,
    progress.position_seconds,
    progress.duration_seconds,
    progress.updated_at
FROM video_watch_progress progress
JOIN home_file file ON file.id = progress.video_id
WHERE progress.client_id = $1
  AND progress.completed = FALSE
  AND progress.position_seconds > 0
  AND file.deleted_at IS NULL
ORDER BY progress.updated_at DESC, file.id DESC
LIMIT $2;
