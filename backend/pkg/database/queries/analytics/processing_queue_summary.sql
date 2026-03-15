SELECT
	COALESCE(SUM(CASE WHEN type = 'metadata' AND status IN ('queued', 'running') THEN 1 ELSE 0 END), 0) AS metadata_pending,
	COALESCE(SUM(CASE WHEN type = 'metadata' AND status = 'failed' THEN 1 ELSE 0 END), 0) AS metadata_failed,
	COALESCE(SUM(CASE WHEN type = 'thumbnail' AND status IN ('queued', 'running') THEN 1 ELSE 0 END), 0) AS thumbnail_pending,
	COALESCE(SUM(CASE WHEN type = 'thumbnail' AND status = 'failed' THEN 1 ELSE 0 END), 0) AS thumbnail_failed
FROM worker_step;
