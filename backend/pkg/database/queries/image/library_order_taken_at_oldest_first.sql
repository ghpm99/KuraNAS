ORDER BY COALESCE(im.taken_at, '-infinity'::timestamptz) ASC, im.file_id ASC
