INSERT INTO
    video_metadata (
        file_id,
        "path",
        format,
        streams,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?) ON CONFLICT (file_id, "path") DO
UPDATE
SET
    format = EXCLUDED.format,
    streams = EXCLUDED.streams RETURNING id,
    created_at;