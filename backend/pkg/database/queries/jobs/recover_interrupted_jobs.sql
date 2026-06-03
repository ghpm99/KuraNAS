-- Jobs left 'running' when the process stopped are orphaned: the scheduler only
-- picks up 'queued' jobs, so they would never be reprocessed. Return them to the
-- queue on startup so their pending steps get another turn.
UPDATE worker_job
SET
    status = 'queued',
    started_at = NULL
WHERE status = 'running';
