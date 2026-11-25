UPDATE captures
SET status = 'promoting',
    failure_reason = ''
WHERE id = $1;
