UPDATE takeout_import
SET state = 'completed',
    total_entries = $2,
    processed_entries = $2,
    imported = $3,
    already_existing = $4,
    ignored = $5,
    failed_entries = $6,
    error = '',
    ended_at = CURRENT_TIMESTAMP
WHERE job_id = $1;
