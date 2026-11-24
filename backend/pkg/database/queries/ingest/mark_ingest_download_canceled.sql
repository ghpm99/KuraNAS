UPDATE ingest_download SET state = 'canceled', speed_bytes_per_second = NULL, eta_seconds = NULL, ended_at = CURRENT_TIMESTAMP
WHERE id = $1 AND state IN ('queued', 'running');
