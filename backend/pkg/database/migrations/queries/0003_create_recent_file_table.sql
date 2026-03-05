CREATE TABLE IF NOT EXISTS
    recent_file (
        id SERIAL PRIMARY KEY,
        ip_address VARCHAR(45) NOT NULL,
        file_id INTEGER NOT NULL,
        accessed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        UNIQUE (ip_address, file_id)
    );