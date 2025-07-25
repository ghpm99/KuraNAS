SELECT
    id,
    file_path,
    format,
    mode,
    width,
    height,
    info,
    created_at
FROM
    image_metadados
WHERE
    id = ?