WITH RECURSIVE descendants AS (
    SELECT id
    FROM jobs
    WHERE id = $1
    UNION ALL
    SELECT j.id
    FROM jobs j
    JOIN descendants d ON j.parent_job_id = d.id
)
UPDATE jobs
SET cancel_requested = true
WHERE
    id IN (SELECT id FROM descendants)
    AND cancel_requested = false
    AND status IN ('queued', 'running');
