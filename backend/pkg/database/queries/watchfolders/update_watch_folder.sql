UPDATE watch_folders
SET path = $2,
    label = $3,
    enabled = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, path, label, enabled, last_scan_at, created_at, updated_at;
