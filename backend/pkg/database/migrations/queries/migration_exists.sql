SELECT
    COUNT(*)
FROM
    migrations
WHERE
    NAME = $1;