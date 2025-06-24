UPDATE home_file
SET
    name = ?,
    "path" = ?,
    parent_path = ?,
    format = ?,
    "size" = ?,
    updated_at = ?,
    created_at = ?,
    last_interaction = ?,
    last_backup = ?,
    "type" = ?,
    checksum = ?,
    deleted_at = ?,
    starred = ?
WHERE
    id = ?;