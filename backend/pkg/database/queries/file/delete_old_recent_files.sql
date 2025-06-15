DELETE FROM recent_file
WHERE
    ip_address = ?
    AND id NOT IN (
        SELECT
            id
        FROM
            recent_file
        WHERE
            ip_address = ?
        ORDER BY
            accessed_at DESC
        LIMIT
            ?
    );