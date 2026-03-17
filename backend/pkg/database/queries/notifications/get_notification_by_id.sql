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
WHERE id = $1;
