-- Package 3: collapse the duplicate-job backlog and enforce one pending job per
-- file from now on.
--
-- The diff bug used to create a fresh job for every file on every scan, so the
-- queue accumulated millions of duplicate pending jobs for only ~30k files. Keep
-- the most recent pending job per file path and cancel the rest, then cancel
-- their still-pending steps so the queue counters reflect reality.
WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY scope ->> 'path'
            ORDER BY created_at DESC, id DESC
        ) AS rn
    FROM worker_job
    WHERE status IN ('queued', 'running')
      AND scope ->> 'path' IS NOT NULL
)
UPDATE worker_job j
SET status = 'canceled',
    ended_at = NOW()
FROM ranked
WHERE j.id = ranked.id
  AND ranked.rn > 1;

-- Cancel any pending step that belongs to a now-canceled job (a canceled job
-- must not leave queued/running steps inflating the pending counters).
UPDATE worker_step s
SET status = 'canceled',
    ended_at = NOW()
FROM worker_job j
WHERE s.job_id = j.id
  AND j.status = 'canceled'
  AND s.status IN ('queued', 'running');

-- From now on, the database itself guarantees at most one pending job per file
-- path. Scan jobs (path NULL) are unaffected. Created after the dedupe above so
-- the index does not fail on existing duplicates.
CREATE UNIQUE INDEX IF NOT EXISTS idx_worker_job_pending_path
    ON worker_job ((scope ->> 'path'))
    WHERE status IN ('queued', 'running') AND scope ->> 'path' IS NOT NULL;
