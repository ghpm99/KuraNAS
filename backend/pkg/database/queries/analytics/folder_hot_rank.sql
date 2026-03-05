SELECT
	parent_path,
	COUNT(*) AS new_files,
	COALESCE(SUM(size), 0) AS added_bytes,
	MAX(created_at) AS last_event_at
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL
	AND created_at >= NOW() - $1::interval
GROUP BY parent_path
ORDER BY added_bytes DESC, new_files DESC
LIMIT $2;
