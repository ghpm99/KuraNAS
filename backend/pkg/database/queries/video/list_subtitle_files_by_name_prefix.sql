SELECT
    id,
    name
FROM home_file
WHERE parent_path = $1
  AND deleted_at IS NULL
  AND LOWER(format) = ANY($2)
  AND LOWER(name) LIKE $3 ESCAPE '\'
ORDER BY name, id;
