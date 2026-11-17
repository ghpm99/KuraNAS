UPDATE takeout_import
SET state = 'failed',
    error = $2,
    ended_at = CURRENT_TIMESTAMP
WHERE job_id = $1;
