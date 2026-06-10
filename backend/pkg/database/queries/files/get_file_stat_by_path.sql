SELECT
    size,
    updated_at
FROM
    home_file
WHERE
    path = $1
    AND deleted_at IS NULL
ORDER BY
    id DESC
LIMIT
    1;
