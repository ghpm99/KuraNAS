INSERT INTO jobs (
    id,
    type,
    priority,
    parent_job_id,
    scope_json,
    status,
    cancel_requested,
    last_error
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
