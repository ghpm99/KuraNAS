SELECT
    description,
    latitude,
    longitude
FROM
    takeout_photo_sidecar
WHERE
    path = $1;
