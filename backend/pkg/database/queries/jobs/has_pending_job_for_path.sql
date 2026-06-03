SELECT EXISTS (
    SELECT 1
    FROM worker_job
    WHERE status IN ('queued', 'running')
      AND scope ->> 'path' = $1
);
