UPDATE worker_job
SET
    status = 'queued',
    next_attempt_at = NOW(),
    started_at = NULL,
    ended_at = NULL
WHERE id = $1;
