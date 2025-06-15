SELECT
    id,
    ip_address,
    file_id,
    accessed_at
FROM
    recent_file
WHERE
    1 = 1
    AND file_id = ?;