-- Exact lookup by name + full path, regardless of soft-delete state. A file
-- recreated at the same path leaves the old soft-deleted row next to the new
-- one, so this can return more than one row; newest first so the caller can
-- prefer the active row.
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
    hf.name = $1
    AND hf.path = $2
ORDER BY
    hf.id DESC
LIMIT
    $3;
