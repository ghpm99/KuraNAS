DELETE FROM image_album_item WHERE album_id = $1 AND file_id = ANY($2)
