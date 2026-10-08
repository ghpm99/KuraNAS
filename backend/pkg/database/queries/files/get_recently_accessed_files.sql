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
    JOIN (
        SELECT
            file_id,
            MAX(accessed_at) AS last_accessed_at
        FROM
            recent_file
        GROUP BY
            file_id
    ) recent ON recent.file_id = hf.id
WHERE
    hf.deleted_at IS NULL
ORDER BY
    recent.last_accessed_at DESC,
    hf.id DESC
LIMIT
    $1
OFFSET
    $2;
