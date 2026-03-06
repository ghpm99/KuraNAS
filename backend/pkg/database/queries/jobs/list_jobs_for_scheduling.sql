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
WHERE status = $1
ORDER BY priority DESC, created_at ASC, id ASC
LIMIT $2;
