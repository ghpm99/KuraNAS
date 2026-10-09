
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($1)
    AND NOT EXISTS (
        SELECT 1 FROM image_metadata newer
        WHERE newer.file_id = im.file_id AND newer.id > im.id
    )
