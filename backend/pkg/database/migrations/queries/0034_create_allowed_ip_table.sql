-- Access control by IP whitelist (registered decision: no authentication).
-- Entries are always stored as CIDR: a single IP becomes /32 (IPv4) or /128
-- (IPv6), so ranges like 192.168.1.0/24 come for free. Loopback is always
-- allowed by the middleware and never needs a row here.
CREATE TABLE IF NOT EXISTS allowed_ip (
    id SERIAL PRIMARY KEY,
    cidr TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
