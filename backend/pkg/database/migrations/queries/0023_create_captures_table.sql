CREATE TABLE IF NOT EXISTS captures (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    file_name VARCHAR(512) NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    media_type VARCHAR(64) NOT NULL DEFAULT '',
    mime_type VARCHAR(128) NOT NULL DEFAULT '',
    size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_captures_name ON captures (name);
CREATE INDEX IF NOT EXISTS idx_captures_media_type ON captures (media_type);
CREATE INDEX IF NOT EXISTS idx_captures_created_at ON captures (created_at DESC);
