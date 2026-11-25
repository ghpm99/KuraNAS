SELECT
	type,
	last_error,
	ended_at
FROM worker_step
WHERE status = 'failed'
	AND ended_at IS NOT NULL
ORDER BY ended_at DESC
LIMIT $1;
