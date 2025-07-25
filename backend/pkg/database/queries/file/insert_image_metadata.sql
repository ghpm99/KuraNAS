INSERT INTO
    image_metadados (
        file_path,
        format,
        mode,
        width,
        height,
        info,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?) RETURNING id,
    created_at