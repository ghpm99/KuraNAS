UPDATE jobs
SET cancel_requested = true
WHERE
    id = $1
    AND cancel_requested = false
    AND status IN ('queued', 'running');
