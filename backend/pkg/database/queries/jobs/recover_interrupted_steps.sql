-- Steps left mid-execution when the process stopped are stuck in 'running' and
-- would never run again. Return them to the queue so the job can finish.
-- Completed steps are untouched, so only the unfinished work reruns.
UPDATE worker_step
SET
    status = 'queued',
    started_at = NULL
WHERE status = 'running';
