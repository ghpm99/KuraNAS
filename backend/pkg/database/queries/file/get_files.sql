SELECT
    hf.id,
    hf.name,
    hf."path",
    hf.parent_path,
    hf.format,
    hf."size",
    hf.updated_at,
    hf.created_at,
    hf.last_interaction,
    hf.last_backup,
    hf."type",
    hf.checksum,
    hf.deleted_at
FROM
    home_file hf
WHERE
    1 = 1
    AND ($1 OR hf.id = $2)
    AND ($3 OR hf.name LIKE '%' || $4 || '%')
    AND ($5 OR hf."path" = $6)
    AND ($7 OR hf."parent_path" = $8)
    AND ($9 OR hf.format = $10)
    AND ($11 OR hf."type" = $12)
    AND ($13 OR hf.deleted_at = $14)
ORDER BY
    type,
    name,
    - id
LIMIT
    $15 OFFSET $16;