SELECT
    id,
    file_id,
    "path",
    mime,
    info,
    tags,
    created_at
FROM
    audio_metadados
WHERE
    id = ?