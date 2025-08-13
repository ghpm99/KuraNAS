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
    hf.starred
FROM
    home_file hf
WHERE
    (
        $1
        OR hf.id = $2
    )
    AND (
        $3
        OR hf.name ILIKE '%' || $4 || '%'
    )
    AND (
        $5
        OR hf.path = $6
    )
    AND (
        $7
        OR hf.parent_path = $8
    )
    AND (
        $9
        OR hf.format = $10
    )
    AND (
        $11
        OR hf.type = $12
    )
    AND (
        $13
        OR hf.deleted_at = $14
    )
    AND CASE $15
        WHEN 'all' THEN TRUE
        WHEN 'recent' THEN hf.id IN (
            SELECT
                file_id
            FROM
                recent_file
        )
        WHEN 'starred' THEN hf.starred = TRUE
        ELSE TRUE
    END
ORDER BY
TYPE,
NAME,
id DESC
LIMIT
    $16
OFFSET
    $17;