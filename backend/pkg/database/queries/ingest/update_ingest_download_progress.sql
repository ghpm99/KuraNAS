UPDATE ingest_download SET progress = $2, speed_bytes_per_second = $3, eta_seconds = $4
WHERE id = $1 AND state = 'running';
