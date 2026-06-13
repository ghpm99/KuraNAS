SELECT id,
       status,
       created_at,
       started_at,
       ended_at,
       COALESCE(last_error, '')
FROM worker_job
WHERE type = 'backup_run'
ORDER BY id DESC
LIMIT 1;
