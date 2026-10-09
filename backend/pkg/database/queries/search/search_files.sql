SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.starred
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.type = 2
    AND lower(hf.name) LIKE $1
    AND lower(hf.name) LIKE ALL ($2::text[])
ORDER BY
    CASE
        WHEN lower(hf.name) = $3 THEN 0
        WHEN lower(hf.name) LIKE $4 THEN 1
        WHEN lower(hf.name) LIKE $5 THEN 2
        ELSE 3
    END,
    hf.starred DESC,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $6;
