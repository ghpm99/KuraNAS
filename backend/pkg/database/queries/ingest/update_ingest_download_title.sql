UPDATE ingest_download SET title = $2
WHERE id = $1 AND state = 'running';
