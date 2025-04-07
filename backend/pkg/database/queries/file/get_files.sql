SELECT
    hf.id,
    hf.name,
    hf."path",
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
    AND hf."path" = $1
ORDER BY
    type,
    name,
    - id
LIMIT
    $2 OFFSET $3;