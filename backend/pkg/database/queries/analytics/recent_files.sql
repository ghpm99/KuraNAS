SELECT
	id,
	name,
	path,
	parent_path,
	size,
	format,
	created_at,
	updated_at
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1;
