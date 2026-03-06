INSERT INTO worker_job (
    type,
    priority,
    scope,
    status,
    cancel_requested,
    last_error
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id, type, priority, scope, status, created_at, started_at, ended_at, cancel_requested, last_error;
