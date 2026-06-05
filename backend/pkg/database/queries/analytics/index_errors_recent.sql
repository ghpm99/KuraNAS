-- Count of step failures in the last 24h, read from the orchestrator
-- (worker_step) instead of the legacy `log` table. Each permanently failed step
-- is one error; this is what the dashboard "errors (24h)" reflects.
SELECT COUNT(*)
FROM worker_step
WHERE status = 'failed'
	AND ended_at >= NOW() - INTERVAL '24 hours';
