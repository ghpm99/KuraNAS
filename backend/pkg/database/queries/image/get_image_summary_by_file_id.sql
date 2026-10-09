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
    focal_length,
    COALESCE(software, ''),
    COALESCE(image_description, ''),
    taken_at,
    CASE
        WHEN gps_latitude IS NULL
        OR gps_longitude IS NULL
        OR (
            gps_latitude = 0
            AND gps_longitude = 0
        ) THEN NULL
        ELSE gps_latitude
    END,
    CASE
        WHEN gps_latitude IS NULL
        OR gps_longitude IS NULL
        OR (
            gps_latitude = 0
            AND gps_longitude = 0
        ) THEN NULL
        ELSE gps_longitude
    END,
    classification_confidence,
    COALESCE(classification_suggested_name, '')
FROM
    image_metadata
WHERE
    file_id = $1
ORDER BY
    id DESC
LIMIT
    1;
