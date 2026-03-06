SELECT
    id,
    job_id,
    type,
    status,
    depends_on_json,
    attempts,
    max_attempts,
    last_error,
    progress,
    payload_json,
    created_at,
    started_at,
    ended_at
FROM steps
WHERE job_id = $1
ORDER BY created_at ASC, id ASC;
