WITH grouped AS (
	SELECT
		checksum,
		MIN(size) AS item_size,
		COUNT(*) AS copies
	FROM home_file
	WHERE type = 2
		AND deleted_at IS NULL
		AND checksum IS NOT NULL
		AND checksum <> ''
	GROUP BY checksum
	HAVING COUNT(*) > 1
)
SELECT
	COUNT(*) AS groups_total,
	COALESCE(SUM(copies), 0) AS files_total,
	COALESCE(SUM((copies - 1) * item_size), 0) AS reclaimable_bytes
FROM grouped;
