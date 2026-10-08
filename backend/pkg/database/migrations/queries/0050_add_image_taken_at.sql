ALTER TABLE image_metadata
ADD COLUMN IF NOT EXISTS taken_at TIMESTAMPTZ NULL;

CREATE FUNCTION pg_temp.parse_exif_timestamp(rawTimestamp TEXT) RETURNS TIMESTAMPTZ AS $$
DECLARE
    parsedTimestamp TIMESTAMPTZ;
BEGIN
    IF rawTimestamp IS NULL OR btrim(rawTimestamp) !~ '^\d{4}:\d{2}:\d{2} \d{2}:\d{2}:\d{2}$' THEN
        RETURN NULL;
    END IF;

    parsedTimestamp := to_timestamp(btrim(rawTimestamp), 'YYYY:MM:DD HH24:MI:SS')::TIMESTAMP AT TIME ZONE 'UTC';

    IF parsedTimestamp < TIMESTAMPTZ '1900-01-01 00:00:00+00' THEN
        RETURN NULL;
    END IF;

    RETURN parsedTimestamp;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

UPDATE image_metadata AS metadata
SET taken_at = COALESCE(
    pg_temp.parse_exif_timestamp(metadata.datetime_original),
    pg_temp.parse_exif_timestamp(metadata.datetime),
    (SELECT file.updated_at FROM home_file AS file WHERE file.id = metadata.file_id)
)
WHERE metadata.taken_at IS NULL;
