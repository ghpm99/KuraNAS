UPDATE watch_folders
SET last_scan_at = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
