UPDATE ingest_download SET state = 'running', progress = 0, speed_bytes_per_second = NULL, eta_seconds = NULL, error = '', ended_at = NULL
WHERE id = $1 AND state IN ('queued', 'running');
