UPDATE worker_job
SET
    status = $1,
    started_at = COALESCE($2, started_at),
    ended_at = COALESCE($3, ended_at),
    cancel_requested = COALESCE($4, cancel_requested),
    last_error = COALESCE($5, last_error)
WHERE id = $6;
