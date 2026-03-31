SELECT id, category, path, created_at, updated_at
FROM libraries
WHERE category = $1;
