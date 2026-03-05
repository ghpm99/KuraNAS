CREATE TABLE IF NOT EXISTS home_file (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    parent_path VARCHAR(1024) NOT NULL,
    format VARCHAR(256) NOT NULL,
    size BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_interaction TIMESTAMPTZ,
    last_backup TIMESTAMPTZ,
    type INTEGER,
    checksum VARCHAR(64),
    deleted_at TIMESTAMPTZ
);