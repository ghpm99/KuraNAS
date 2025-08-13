INSERT INTO
    activity_diary (NAME, description, start_time)
VALUES
    ($1, $2, $3)
RETURNING
    id;