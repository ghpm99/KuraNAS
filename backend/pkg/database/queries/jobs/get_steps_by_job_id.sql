SELECT
    id,
    job_id,
    type,
    status,
    depends_on,
    attempts,
    max_attempts,
    last_error,
    progress,
    payload,
    created_at,
    started_at,
    ended_at
FROM worker_step
WHERE job_id = $1
ORDER BY created_at ASC, id ASC;
