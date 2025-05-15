SELECT
    id,
    name,
    description,
    start_time,
    end_time
FROM
    activity_diary
WHERE
    1 = 1
    AND ($1 OR id = $2)
    AND ($3 OR name LIKE '%' || $4 || '%')
    AND ($5 OR description LIKE '%' || $6 || '%')
    AND ($7 OR start_time = $8)
    AND ($9 OR end_time = $10)
ORDER BY
	- start_time,
    - id
LIMIT
    $11 OFFSET $12;