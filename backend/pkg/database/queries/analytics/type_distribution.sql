SELECT
	CASE
		WHEN lower(COALESCE(format, '')) = ANY($1::text[]) THEN 'video'
		WHEN lower(COALESCE(format, '')) = ANY($2::text[]) THEN 'image'
		WHEN lower(COALESCE(format, '')) = ANY($3::text[]) THEN 'audio'
		WHEN lower(COALESCE(format, '')) = ANY($4::text[]) THEN 'document'
		WHEN lower(COALESCE(format, '')) = ANY($5::text[]) THEN 'archive'
		ELSE 'other'
	END AS category,
	COUNT(*) AS total_count,
	COALESCE(SUM(size), 0) AS total_bytes
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL
GROUP BY category
ORDER BY total_bytes DESC;
