CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    type VARCHAR(40) NOT NULL CHECK (type IN ('startup_scan', 'upload_process', 'fs_event', 'reindex_folder')),
    priority INTEGER NOT NULL CHECK (priority BETWEEN 1 AND 4),
    scope_json TEXT NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL CHECK (status IN ('queued', 'running', 'partial_fail', 'failed', 'completed', 'canceled')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    cancel_requested BOOLEAN NOT NULL DEFAULT FALSE,
    last_error TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS steps (
    id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    type VARCHAR(40) NOT NULL CHECK (type IN ('scan_filesystem', 'diff_against_db', 'metadata', 'checksum', 'persist', 'thumbnail', 'playlist_index', 'mark_deleted')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed', 'canceled', 'skipped')),
    depends_on_json TEXT NOT NULL DEFAULT '[]',
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    max_attempts INTEGER NOT NULL DEFAULT 1 CHECK (max_attempts >= 1),
    last_error TEXT NOT NULL DEFAULT '',
    progress INTEGER NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    payload_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_priority_created_at
    ON jobs (status, priority DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_jobs_created_at
    ON jobs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_steps_job_id_created_at
    ON steps (job_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_steps_job_id_status
    ON steps (job_id, status);
