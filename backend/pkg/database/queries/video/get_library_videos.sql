SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at
FROM
    home_file hf
WHERE
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
    AND (
        $2 = ''
        OR LOWER(hf.name) LIKE LOWER($2)
        OR LOWER(hf.path) LIKE LOWER($2)
        OR LOWER(hf.parent_path) LIKE LOWER($2)
        OR LOWER(hf.format) LIKE LOWER($2)
    )
ORDER BY
    hf.updated_at DESC,
    hf.id DESC
LIMIT
    $3
OFFSET
    $4;
