SELECT
    id,
    type,
    priority,
    scope,
    status,
    created_at,
    started_at,
    ended_at,
    cancel_requested,
    last_error
FROM worker_job
WHERE id = $1;
