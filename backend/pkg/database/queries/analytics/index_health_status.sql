SELECT
	status,
	started_at,
	ended_at
FROM worker_job
WHERE type IN ('startup_scan', 'reindex_folder')
ORDER BY created_at DESC
LIMIT 1;
