SELECT
    CASE
        WHEN duration ~ '^[0-9]+(\.[0-9]+)?$' THEN duration::float8
    END
FROM
    video_metadata
WHERE
    file_id = $1
ORDER BY
    id DESC
LIMIT
    1;
