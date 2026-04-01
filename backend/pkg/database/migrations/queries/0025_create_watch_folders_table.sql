CREATE TABLE IF NOT EXISTS watch_folders (
    id           SERIAL PRIMARY KEY,
    path         VARCHAR(500) NOT NULL UNIQUE,
    label        VARCHAR(100),
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    last_scan_at TIMESTAMP,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
