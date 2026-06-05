-- Most recent step failures, read from the orchestrator (worker_step). The step
-- type stands in for the error "name" and last_error for the description; the
-- full stack lives in the forensic file log. Replaces the legacy `log` table,
-- which the orchestrator no longer writes to.
SELECT
	type,
	last_error,
	ended_at
FROM worker_step
WHERE status = 'failed'
	AND ended_at IS NOT NULL
ORDER BY ended_at DESC
LIMIT $1;
