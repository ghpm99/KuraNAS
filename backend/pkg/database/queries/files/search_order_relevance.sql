ORDER BY
    CASE
        WHEN lower(hf.name) = @1 THEN 0
        WHEN lower(hf.name) LIKE @2 THEN 1
        WHEN lower(hf.name) LIKE @3 THEN 2
        ELSE 3
    END,
    hf.starred DESC,
    hf.updated_at DESC,
    hf.id ASC
