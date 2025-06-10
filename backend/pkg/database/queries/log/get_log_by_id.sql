SELECT
    id,
    name,
    description,
    level,
    ip_address,
    start_time,
    end_time,
    created_at,
    updated_at,
    deleted_at,
    status,
    extra_data
FROM
    log
WHERE
    id = ?;