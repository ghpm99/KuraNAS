SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($2)
    AND (
        hf.name ILIKE '%' || $1 || '%'
        OR hf.path ILIKE '%' || $1 || '%'
    )
ORDER BY
    CASE
        WHEN LOWER(hf.name) = LOWER($1) THEN 0
        WHEN hf.name ILIKE $1 || '%' THEN 1
        ELSE 2
    END,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $3;
