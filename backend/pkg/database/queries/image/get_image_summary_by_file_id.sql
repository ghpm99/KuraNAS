SELECT
    width,
    height,
    make,
    model,
    lens_model,
    datetime_original,
    exposure_time,
    f_number,
    iso,
    focal_length
FROM
    image_metadata
WHERE
    file_id = $1
ORDER BY
    id DESC
LIMIT
    1;
