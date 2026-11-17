SELECT id
FROM worker_job
WHERE status IN ('queued', 'running')
  AND scope ->> 'reindex_key' = $1
ORDER BY id
LIMIT 1;
