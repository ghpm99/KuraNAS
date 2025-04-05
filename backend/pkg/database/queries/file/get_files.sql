SELECT
    hf.id,
    hf.name,
    hf."path",
    hf.format,
    hf."size",
    hf.updated_at,
    hf.created_at,
    hf.last_interaction,
    hf.last_backup
FROM
    home_file hf
WHERE
    1 = 1
    AND ($3 IS '' OR hf.name LIKE '%' || $3 || '%')
    AND ($4 IS '' OR hf."path" LIKE '%' || $4 || '%')
ORDER BY
    - id
LIMIT
    $1 OFFSET $2;