SELECT
	id,
    name,
    description,
    start_time,
    end_time
FROM
	activity_diary ad
WHERE
	1 = 1
	AND ad.start_time BETWEEN $1 AND $2
	OR ad.end_time BETWEEN $1 AND $2
	OR ad.end_time IS NULL
ORDER BY
	- start_time,
    - id