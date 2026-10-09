SELECT
    EXTRACT(YEAR FROM im.taken_at AT TIME ZONE 'UTC')::int AS taken_year,
    EXTRACT(MONTH FROM im.taken_at AT TIME ZONE 'UTC')::int AS taken_month,
    count(*)
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
