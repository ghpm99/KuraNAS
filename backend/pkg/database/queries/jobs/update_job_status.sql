UPDATE jobs
SET
    status = $3,
    started_at = COALESCE($4, started_at),
    ended_at = $5,
    last_error = $6
WHERE
    id = $1
    AND status = $2;
