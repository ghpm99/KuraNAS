SELECT
    id,
    file_id,
    "path",
    mime,
    info,
    tags,
    created_at
FROM
    audio_metadata
WHERE
    id = ?