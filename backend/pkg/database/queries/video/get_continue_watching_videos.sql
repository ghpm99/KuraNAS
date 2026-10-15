SELECT
    file.id,
    file.name,
    file.path,
    file.parent_path,
    file.format,
    file.size,
    file.created_at,
    file.updated_at,
    COALESCE(file.starred, FALSE),
    tech.duration_seconds,
    tech.width,
    tech.height,
    tech.codec_name,
    progress.position_seconds,
    progress.duration_seconds,
    progress.updated_at
FROM video_watch_progress progress
JOIN home_file file ON file.id = progress.video_id
LEFT JOIN LATERAL (
    SELECT
        CASE
            WHEN vm.duration ~ '^[0-9]+(\.[0-9]+)?$' THEN vm.duration::double precision
        END AS duration_seconds,
        vm.width,
        vm.height,
        vm.codec_name
    FROM
        video_metadata vm
    WHERE
        vm.id = (
            SELECT
                MAX(latest.id)
            FROM
                video_metadata latest
            WHERE
                latest.file_id = file.id
        )
) tech ON TRUE
WHERE progress.client_id = $1
  AND progress.completed = FALSE
  AND progress.position_seconds > 0
  AND file.deleted_at IS NULL
ORDER BY progress.updated_at DESC, file.id DESC
LIMIT $2;
