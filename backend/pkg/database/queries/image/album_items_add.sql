INSERT INTO image_album_item (album_id, file_id)
SELECT $1, hf.id
FROM home_file hf
WHERE hf.id = ANY($2)
    AND hf.deleted_at IS NULL
    AND hf.format = ANY($3)
    AND EXISTS (SELECT 1 FROM image_metadata im WHERE im.file_id = hf.id)
ON CONFLICT DO NOTHING
