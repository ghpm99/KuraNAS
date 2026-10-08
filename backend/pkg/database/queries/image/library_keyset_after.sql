((COALESCE(im.taken_at, '-infinity'::timestamptz), im.file_id) < (@1::timestamptz, @2))
