UPDATE takeout_import
SET total_entries = $2,
    processed_entries = $3
WHERE job_id = $1;
