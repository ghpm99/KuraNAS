-- Active row(s) at an exact path, paginated for response-shape compatibility
-- (path is unique among active rows in practice).
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
    hf.path = $1
    AND hf.deleted_at IS NULL
ORDER BY
    hf.type,
    hf.name,
    hf.id DESC
LIMIT
    $2
OFFSET
    $3;
