CREATE TABLE IF NOT EXISTS worker_job (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL CHECK (priority IN ('low', 'normal', 'high')),
    scope JSON,
    status VARCHAR(30) NOT NULL CHECK (status IN ('queued', 'running', 'partial_fail', 'failed', 'completed', 'canceled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    cancel_requested BOOLEAN NOT NULL DEFAULT FALSE,
    last_error TEXT
);

CREATE TABLE IF NOT EXISTS worker_step (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES worker_job(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed', 'canceled', 'skipped')),
    depends_on JSON,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    last_error TEXT,
    progress INTEGER NOT NULL DEFAULT 0,
    payload JSON,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_worker_job_status_priority_created_at
    ON worker_job(status, priority, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_worker_job_created_at
    ON worker_job(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_worker_step_job_id_status
    ON worker_step(job_id, status);

CREATE INDEX IF NOT EXISTS idx_worker_step_job_id_created_at
    ON worker_step(job_id, created_at ASC);
