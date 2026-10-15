SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.updated_at,
    hf.created_at,
    hf.last_interaction,
    hf.last_backup,
    hf.type,
    hf.checksum,
    hf.deleted_at,
    hf.starred,
    hf.physical_path
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.starred = TRUE
    AND CASE $1::text
        WHEN 'folders' THEN hf.type = 1
        WHEN 'files' THEN hf.type = 2 AND NOT (lower(hf.format) = ANY ($2::text[]))
        WHEN 'media' THEN hf.type = 2 AND lower(hf.format) = ANY ($2::text[])
        ELSE TRUE
    END
ORDER BY
    hf.type,
    hf.name,
    hf.id DESC
LIMIT
    $3
OFFSET
    $4;
