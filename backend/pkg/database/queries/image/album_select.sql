SELECT
    a.id,
    a.name,
    (
        SELECT count(*)
        FROM image_album_item ai
        JOIN home_file hf ON hf.id = ai.file_id
        WHERE ai.album_id = a.id AND hf.deleted_at IS NULL
    ) AS item_count,
    COALESCE(
        (
            SELECT ai.file_id
            FROM image_album_item ai
            JOIN home_file hf ON hf.id = ai.file_id
            WHERE ai.album_id = a.id AND ai.file_id = a.cover_file_id AND hf.deleted_at IS NULL
        ),
        (
            SELECT ai.file_id
            FROM image_album_item ai
            JOIN home_file hf ON hf.id = ai.file_id
            WHERE ai.album_id = a.id AND hf.deleted_at IS NULL
            ORDER BY ai.added_at DESC, ai.file_id DESC
            LIMIT 1
        )
    ) AS cover_file_id,
    a.created_at,
    a.updated_at
FROM image_album a
