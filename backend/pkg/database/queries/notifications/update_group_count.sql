UPDATE notifications
SET group_count = $1, message = $2
WHERE id = $3;
