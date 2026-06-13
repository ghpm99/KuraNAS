UPDATE email_message
SET status = $2, prefilter_rules = $3
WHERE id = $1;
