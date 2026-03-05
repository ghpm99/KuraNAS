SELECT
	parent_path,
	COUNT(*) AS total_files,
	COALESCE(SUM(size), 0) AS total_bytes,
	MAX(updated_at) AS last_modified_at
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL
GROUP BY parent_path
ORDER BY total_bytes DESC
LIMIT $1;
