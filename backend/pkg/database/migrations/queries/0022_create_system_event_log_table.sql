CREATE TABLE IF NOT EXISTS system_event_log (
    id BIGSERIAL PRIMARY KEY,
    event_time TIMESTAMPTZ NOT NULL,
    event_time_display VARCHAR(19) NOT NULL,
    event_type VARCHAR(32) NOT NULL CHECK (event_type IN ('STARTUP', 'SHUTDOWN')),
    description TEXT NOT NULL,
    source VARCHAR(64) NOT NULL DEFAULT 'backend',
    host_name VARCHAR(255),
    process_id INTEGER CHECK (process_id >= 0),
    extra_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_system_event_log_event_time
    ON system_event_log (event_time DESC);
