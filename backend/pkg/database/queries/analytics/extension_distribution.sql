SELECT
	COALESCE(NULLIF(lower(format), ''), '<none>') AS extension,
	COUNT(*) AS total_count,
	COALESCE(SUM(size), 0) AS total_bytes
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL
GROUP BY extension
ORDER BY total_bytes DESC
LIMIT $1;
