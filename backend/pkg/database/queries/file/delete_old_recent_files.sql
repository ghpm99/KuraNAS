DELETE FROM recent_file
WHERE
    ip_address = $1
    AND id NOT IN (
        SELECT
            id
        FROM
            recent_file
        WHERE
            ip_address = $2
        ORDER BY
            accessed_at DESC
        LIMIT
            $3
    );