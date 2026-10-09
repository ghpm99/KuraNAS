lower(hf.name) LIKE @1
    AND lower(hf.name) LIKE ALL (@2::text[])
