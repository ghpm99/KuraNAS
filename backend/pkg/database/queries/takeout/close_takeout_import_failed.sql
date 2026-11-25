UPDATE takeout_import
SET state = 'failed',
    error = $2,
    total_entries = $3,
    processed_entries = $4::int + $5::int + $6::int + $7::int,
    imported = $4,
    already_existing = $5,
    ignored = $6,
    failed_entries = $7,
    ended_at = CURRENT_TIMESTAMP
WHERE job_id = $1;
