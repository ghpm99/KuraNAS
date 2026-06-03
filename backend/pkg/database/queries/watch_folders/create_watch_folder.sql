INSERT INTO watch_folders (path, label, enabled, last_scan_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, path, label, enabled, last_scan_at, created_at, updated_at;
