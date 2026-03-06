UPDATE worker_step
SET
    status = $1,
    progress = $2,
    attempts = $3,
    started_at = COALESCE($4, started_at),
    ended_at = COALESCE($5, ended_at),
    last_error = COALESCE($6, last_error)
WHERE id = $7;
