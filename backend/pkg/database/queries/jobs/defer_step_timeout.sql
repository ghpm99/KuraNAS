UPDATE worker_step
SET
    status = 'queued',
    attempts = $2,
    timeout_count = timeout_count + 1,
    started_at = NULL,
    last_error = $3
WHERE id = $1;
