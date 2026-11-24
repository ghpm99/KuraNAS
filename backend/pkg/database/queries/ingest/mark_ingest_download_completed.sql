UPDATE ingest_download SET state = 'completed', progress = 100, speed_bytes_per_second = NULL, eta_seconds = NULL, title = $2, final_path = $3, error = '', ended_at = CURRENT_TIMESTAMP
WHERE id = $1 AND state IN ('queued', 'running');
