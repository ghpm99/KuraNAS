SELECT
    id,
    name,
    path,
    parent_path,
    format,
    size,
    created_at,
    updated_at
FROM home_file
WHERE deleted_at IS NULL
  AND format = ANY($1)
ORDER BY updated_at DESC, id DESC
LIMIT $2;
