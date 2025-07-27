SELECT
    id,
    file_id,
    "path",
    format,
    mode,
    width,
    height,
    info,
    created_at
FROM
    image_metadata
WHERE
    id = ?