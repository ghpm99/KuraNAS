UPDATE activity_diary
SET
    NAME = $1,
    description = $2,
    start_time = $3,
    end_time = $4
WHERE
    id = $5;