INSERT INTO jobs (
    id,
    type,
    priority,
    scope_json,
    status,
    cancel_requested,
    last_error
)
VALUES ($1, $2, $3, $4, $5, $6, $7);
