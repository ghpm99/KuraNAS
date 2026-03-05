WITH days AS (
	SELECT generate_series(
		date_trunc('day', NOW() - $1::interval),
		date_trunc('day', NOW()),
		'1 day'::interval
	) AS day
)
SELECT
	day::date,
	COALESCE((
		SELECT SUM(hf.size)
		FROM home_file hf
		WHERE hf.type = 2
			AND hf.created_at <= (day + '1 day'::interval)
			AND (hf.deleted_at IS NULL OR hf.deleted_at > day)
	), 0) AS used_bytes
FROM days
ORDER BY day;
