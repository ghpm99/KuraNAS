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
WHERE
    ($1::boolean OR status = $2)
    AND ($3::boolean OR type = $4)
    AND ($5::boolean OR priority = $6)
ORDER BY created_at DESC
LIMIT $7 OFFSET $8;
