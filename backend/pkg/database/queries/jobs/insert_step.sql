INSERT INTO worker_step (
    job_id,
    type,
    status,
    depends_on,
    attempts,
    max_attempts,
    last_error,
    progress,
    payload
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
RETURNING id, job_id, type, status, depends_on, attempts, max_attempts, last_error, progress, payload, created_at, started_at, ended_at;
