CREATE TABLE IF NOT EXISTS ingest_download (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES worker_job(id) ON DELETE SET NULL,
    url TEXT NOT NULL,
    title TEXT,
    preset VARCHAR(64) NOT NULL,
    target_root_id INTEGER NOT NULL,
    target_root_name TEXT NOT NULL,
    subfolder TEXT NOT NULL DEFAULT '',
    state VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (state IN ('queued', 'running', 'completed', 'failed', 'canceled')),
    progress INTEGER NOT NULL DEFAULT 0,
    speed_bytes_per_second BIGINT,
    eta_seconds INTEGER,
    file_id INTEGER,
    error TEXT NOT NULL DEFAULT '',
    final_path TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ingest_download_created_at_id ON ingest_download (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_ingest_download_state ON ingest_download (state);
CREATE INDEX IF NOT EXISTS idx_ingest_download_job_id ON ingest_download (job_id);
