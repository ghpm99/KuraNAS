-- All active files, paginated — the flat /files listing.
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
    hf.deleted_at IS NULL
ORDER BY
    hf.type,
    hf.name,
    hf.id DESC
LIMIT
    $1
OFFSET
    $2;
