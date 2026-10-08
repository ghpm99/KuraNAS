ORDER BY COALESCE(im.taken_at, '-infinity'::timestamptz) DESC, im.file_id DESC
