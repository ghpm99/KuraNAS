hf.id IN (SELECT album_item.file_id FROM image_album_item album_item WHERE album_item.album_id = @1)
