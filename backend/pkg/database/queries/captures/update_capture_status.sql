UPDATE captures
SET status = $2,
    file_id = $3
WHERE id = $1;
