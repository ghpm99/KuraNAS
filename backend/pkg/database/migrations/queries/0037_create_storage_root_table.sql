-- Storage roots: the N directories KuraNAS indexes, watches and serves.
-- Seeded at boot from ENTRY_POINT when empty, so existing installs migrate
-- without user action. Paths are absolute and unique.
CREATE TABLE IF NOT EXISTS storage_root (
    id SERIAL PRIMARY KEY,
    path VARCHAR(1024) NOT NULL UNIQUE,
    label VARCHAR(256) NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
