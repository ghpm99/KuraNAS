-- Active recently-accessed children of a directory, paginated — tree listing
-- with the "recent" category.
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
    hf.parent_path = $1
    AND hf.deleted_at IS NULL
    AND hf.id IN (
        SELECT
            file_id
        FROM
            recent_file
    )
ORDER BY
    hf.type,
    hf.name,
    hf.id DESC
LIMIT
    $2
OFFSET
    $3;
