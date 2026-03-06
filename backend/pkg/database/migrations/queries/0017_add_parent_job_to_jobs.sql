ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS parent_job_id TEXT NULL REFERENCES jobs(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_jobs_parent_job_id
    ON jobs (parent_job_id);
