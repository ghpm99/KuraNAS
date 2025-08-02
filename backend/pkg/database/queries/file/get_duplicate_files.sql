SELECT
    name,
    SUM(size) AS total_size,
    COUNT(*) AS copies,
    GROUP_CONCAT(path) AS paths
FROM
    home_file
WHERE
    checksum IN (
        SELECT
            checksum
        FROM
            home_file
        WHERE
            checksum IS NOT ''
        GROUP BY
            checksum
        HAVING
            COUNT(*) > 1
    )
GROUP BY
    checksum
ORDER BY
    copies DESC
LIMIT
    ? OFFSET ?;