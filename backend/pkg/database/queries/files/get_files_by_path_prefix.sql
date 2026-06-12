-- Paginated walk of a subtree (root row included), regardless of soft-delete
-- state — used by the mark_deleted reconciliation, which both soft-deletes
-- missing files and revives rows whose file reappeared on disk.
-- Literal prefix operator ^@ (starts_with), never LIKE: PostgreSQL treats '\'
-- as the LIKE escape character, so Windows paths (e.g. D:\Folder) would
-- silently match nothing. The operator form is served by the SP-GiST index
-- home_file_path_prefix; a bare starts_with() call is not.
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
    hf.path ^@ $1
ORDER BY
    hf.type,
    hf.name,
    hf.id DESC
LIMIT
    $2
OFFSET
    $3;
