WITH grouped AS (
	SELECT
		checksum,
		MIN(size) AS item_size,
		COUNT(*) AS copies,
		jsonb_agg(
			jsonb_build_object(
				'id', id,
				'path', path,
				'updated_at', updated_at,
				'tier', CASE WHEN physical_path IS NULL THEN 'hot' ELSE 'cold' END
			)
			ORDER BY path
		) AS copy_files
	FROM home_file
	WHERE type = 2
		AND deleted_at IS NULL
		AND checksum IS NOT NULL
		AND checksum <> ''
		AND ($1::text[] IS NULL OR lower(COALESCE(format, '')) = ANY($1::text[]))
		AND NOT (lower(COALESCE(format, '')) = ANY($2::text[]))
	GROUP BY checksum
	HAVING COUNT(*) > 1
)
SELECT
	checksum,
	copies,
	item_size,
	(copies - 1) * item_size AS reclaimable_bytes,
	copy_files
FROM grouped
ORDER BY reclaimable_bytes DESC, checksum ASC
LIMIT $3 OFFSET $4;
