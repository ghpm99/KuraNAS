INSERT INTO ingest_download (job_id, url, preset, target_root_id, target_root_name, subfolder, state)
VALUES ($1, $2, $3, $4, $5, $6, 'queued')
RETURNING id;
