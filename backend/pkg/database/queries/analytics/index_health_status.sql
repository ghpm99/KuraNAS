-- Health of the most recent scan, read from the job orchestrator (worker_job)
-- which is the source of truth for the current pipeline. The legacy `log` table
-- is no longer written by the orchestrator, so reading it showed stale/empty
-- health. A "scan" is a full startup scan or a folder reindex job.
SELECT
	status,
	started_at,
	ended_at
FROM worker_job
WHERE type IN ('startup_scan', 'reindex_folder')
ORDER BY created_at DESC
LIMIT 1;
