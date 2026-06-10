SELECT
    NAME,
    SUM(size) AS total_size,
    COUNT(*) AS copies,
    STRING_AGG(PATH, ',') AS paths
FROM
    home_file
WHERE
    checksum IN (
        SELECT
            checksum
        FROM
            home_file
        WHERE
            checksum IS NOT NULL
            AND checksum <> ''
        GROUP BY
            checksum
        HAVING
            COUNT(*) > 1
    )
GROUP BY
    checksum,
    NAME
ORDER BY
    copies DESC
LIMIT
    $1
OFFSET
    $2;