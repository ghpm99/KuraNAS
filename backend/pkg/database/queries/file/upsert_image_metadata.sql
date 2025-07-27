INSERT INTO
    image_metadata (
        file_id,
        "path",
        format,
        mode,
        width,
        height,
        info,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (file_id, "path") DO
UPDATE
SET
    format = EXCLUDED.format,
    mode = EXCLUDED.mode,
    width = EXCLUDED.width,
    height = EXCLUDED.height,
    info = EXCLUDED.info RETURNING id,
    created_at;