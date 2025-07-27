SELECT
    id,
    file_id,
    "path",
    format,
    streams,
    created_at
FROM
    video_metadata
WHERE
    id = ?