INSERT INTO
    takeout_photo_sidecar (path, description, latitude, longitude)
VALUES
    ($1, $2, $3, $4) ON CONFLICT (path)
DO
UPDATE
SET
    description = EXCLUDED.description,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude;
