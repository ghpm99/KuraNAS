SELECT
    id,
    name
FROM home_file
WHERE parent_path = $1
  AND deleted_at IS NULL
  AND LOWER(format) = ANY($2)
ORDER BY id;
