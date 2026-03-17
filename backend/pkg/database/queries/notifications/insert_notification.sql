INSERT INTO notifications (
    type,
    title,
    message,
    metadata,
    is_read,
    group_key,
    group_count,
    is_grouped
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING id, type, title, message, metadata, is_read, created_at, group_key, group_count, is_grouped;
