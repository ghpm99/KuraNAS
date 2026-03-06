SELECT
    id,
    type,
    priority,
    scope_json,
    status,
    created_at,
    started_at,
    ended_at,
    cancel_requested,
    last_error
FROM jobs
WHERE id = $1;
