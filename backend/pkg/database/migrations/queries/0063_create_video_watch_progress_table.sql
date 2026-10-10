CREATE TABLE IF NOT EXISTS video_watch_progress (
    client_id VARCHAR(128) NOT NULL,
    video_id INTEGER NOT NULL REFERENCES home_file(id) ON DELETE CASCADE,
    position_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (client_id, video_id)
);

CREATE INDEX IF NOT EXISTS idx_video_watch_progress_client_updated_at
    ON video_watch_progress (client_id, updated_at DESC);
