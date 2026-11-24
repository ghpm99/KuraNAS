UPDATE ingest_download SET state = 'failed', speed_bytes_per_second = NULL, eta_seconds = NULL, error = $2, ended_at = CURRENT_TIMESTAMP
WHERE id = $1 AND state IN ('queued', 'running');
