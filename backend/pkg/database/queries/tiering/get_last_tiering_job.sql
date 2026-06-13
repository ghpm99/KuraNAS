SELECT id,
       status,
       created_at,
       started_at,
       ended_at,
       COALESCE(last_error, '')
FROM worker_job
WHERE type = 'tier_migration'
ORDER BY id DESC
LIMIT 1;
