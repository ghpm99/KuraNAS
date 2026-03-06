SELECT
    id,
    type,
    priority,
    scope_json,
    status,
    created_at,
    started_at,
    ended_at,
    cancel_requested,
    last_error
FROM jobs
WHERE
    ($1 OR status = $2)
    AND ($3 OR type = $4)
    AND ($5 OR priority = $6)
ORDER BY created_at DESC, id DESC
LIMIT $7
OFFSET $8;
