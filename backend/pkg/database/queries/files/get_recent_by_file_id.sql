SELECT
    id,
    ip_address,
    file_id,
    accessed_at
FROM
    recent_file
WHERE
    file_id = $1;