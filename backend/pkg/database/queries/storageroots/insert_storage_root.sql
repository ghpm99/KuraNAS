-- Registers a storage root.
INSERT INTO storage_root (path, label, enabled)
VALUES ($1, $2, $3)
RETURNING id, path, label, enabled, created_at;
