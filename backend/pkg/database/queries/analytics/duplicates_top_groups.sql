WITH grouped AS (
	SELECT
		checksum,
		MIN(size) AS item_size,
		COUNT(*) AS copies,
		ARRAY_AGG(path ORDER BY path) AS paths
	FROM home_file
	WHERE type = 2
		AND deleted_at IS NULL
		AND checksum IS NOT NULL
		AND checksum <> ''
	GROUP BY checksum
	HAVING COUNT(*) > 1
)
SELECT
	checksum,
	copies,
	item_size,
	(copies - 1) * item_size AS reclaimable_bytes,
	paths
FROM grouped
ORDER BY reclaimable_bytes DESC
LIMIT $1;
