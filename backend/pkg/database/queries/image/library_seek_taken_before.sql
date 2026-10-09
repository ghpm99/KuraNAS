(COALESCE(im.taken_at, '-infinity'::timestamptz) < @1::timestamptz)
