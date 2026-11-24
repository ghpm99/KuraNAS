UPDATE captures
SET status = 'failed',
    file_id = NULL,
    file_name = $2,
    file_path = $3,
    failure_reason = $4
WHERE id = $1;
