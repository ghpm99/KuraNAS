SELECT COUNT(*)
FROM worker_step
WHERE status = 'failed'
	AND ended_at >= NOW() - INTERVAL '24 hours';
