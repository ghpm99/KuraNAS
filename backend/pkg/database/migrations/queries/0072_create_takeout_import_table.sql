CREATE TABLE IF NOT EXISTS takeout_import (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES worker_job(id) ON DELETE SET NULL,
    file_name TEXT NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    state VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (state IN ('queued', 'running', 'completed', 'failed')),
    total_entries INTEGER NOT NULL DEFAULT 0,
    processed_entries INTEGER NOT NULL DEFAULT 0,
    imported INTEGER NOT NULL DEFAULT 0,
    already_existing INTEGER NOT NULL DEFAULT 0,
    ignored INTEGER NOT NULL DEFAULT 0,
    failed_entries INTEGER NOT NULL DEFAULT 0,
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_takeout_import_created_at_id ON takeout_import (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_takeout_import_job_id ON takeout_import (job_id);
