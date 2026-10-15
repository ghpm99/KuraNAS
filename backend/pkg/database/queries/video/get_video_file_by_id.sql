SELECT
    id,
    name,
    path,
    parent_path,
    format,
    size,
    created_at,
    updated_at,
    COALESCE(starred, FALSE)
FROM home_file
WHERE id = $1
  AND deleted_at IS NULL
  AND format = ANY($2)
LIMIT 1;
