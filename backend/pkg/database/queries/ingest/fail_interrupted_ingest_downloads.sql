WITH interrupted_downloads AS (
    UPDATE ingest_download
    SET state = 'failed', speed_bytes_per_second = NULL, eta_seconds = NULL, error = $1, ended_at = CURRENT_TIMESTAMP
    WHERE state IN ('queued', 'running')
    RETURNING job_id
),
failed_steps AS (
    UPDATE worker_step
    SET status = 'failed', last_error = $1, ended_at = CURRENT_TIMESTAMP
    WHERE status IN ('queued', 'running')
      AND job_id IN (SELECT job_id FROM interrupted_downloads)
),
failed_jobs AS (
    UPDATE worker_job
    SET status = 'failed', last_error = $1, ended_at = CURRENT_TIMESTAMP
    WHERE status IN ('queued', 'running')
      AND id IN (SELECT job_id FROM interrupted_downloads)
)
SELECT COUNT(*) FROM interrupted_downloads;
