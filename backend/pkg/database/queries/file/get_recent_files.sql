SELECT
    file_id,
    accessed_at
FROM
    recent_file
ORDER BY
    accessed_at DESC
LIMIT
    ?
OFFSET
    ?;