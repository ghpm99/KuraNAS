ORDER BY
    CASE
        WHEN kuranas_fold(hf.name) = kuranas_fold(@1) THEN 0
        WHEN kuranas_fold(hf.name) LIKE kuranas_fold(@2) THEN 1
        WHEN kuranas_fold(hf.name) LIKE kuranas_fold(@3) THEN 2
        ELSE 3
    END,
    hf.starred DESC,
    hf.updated_at DESC,
    hf.id ASC
