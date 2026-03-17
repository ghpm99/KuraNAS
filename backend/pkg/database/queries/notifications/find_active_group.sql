SELECT
    id,
    type,
    title,
    message,
    metadata,
    is_read,
    created_at,
    group_key,
    group_count,
    is_grouped
FROM notifications
WHERE
    group_key = $1
    AND type = $2
    AND created_at >= NOW() - ($3 || ' seconds')::INTERVAL
ORDER BY created_at DESC
LIMIT 1;
