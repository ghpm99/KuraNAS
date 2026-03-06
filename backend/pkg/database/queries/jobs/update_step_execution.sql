UPDATE steps
SET
    attempts = $2,
    last_error = $3,
    progress = $4,
    started_at = COALESCE($5, started_at),
    ended_at = $6
WHERE id = $1;
