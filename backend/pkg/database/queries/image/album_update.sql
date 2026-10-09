UPDATE image_album
SET
    name = COALESCE($2::text, name),
    cover_file_id = COALESCE($3::integer, cover_file_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND (
        $3::integer IS NULL
        OR EXISTS (SELECT 1 FROM image_album_item WHERE album_id = $1 AND file_id = $3::integer)
    )
