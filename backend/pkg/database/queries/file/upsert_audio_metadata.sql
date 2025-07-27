INSERT INTO
    audio_metadados (
        file_id,
        "path",
        mime,
        info,
        tags,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?, ?) ON CONFLICT (file_id, "path") DO
UPDATE
SET
    mime = EXCLUDED.mime,
    info = EXCLUDED.info,
    tags = EXCLUDED.tags RETURNING id,
    created_at;