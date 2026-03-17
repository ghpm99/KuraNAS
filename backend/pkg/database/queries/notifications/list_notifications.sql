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
    ($1::boolean OR type = $2)
    AND ($3::boolean OR is_read = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;
