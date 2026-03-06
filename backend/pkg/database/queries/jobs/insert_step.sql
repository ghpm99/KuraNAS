INSERT INTO steps (
    id,
    job_id,
    type,
    status,
    depends_on_json,
    attempts,
    max_attempts,
    last_error,
    progress,
    payload_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
