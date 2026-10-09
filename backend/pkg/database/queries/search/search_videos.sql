SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.updated_at
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($6)
    AND kuranas_fold(hf.name) LIKE kuranas_fold($1)
    AND kuranas_fold(hf.name) LIKE ALL (kuranas_fold_terms($2::text[]))
ORDER BY
    CASE
        WHEN kuranas_fold(hf.name) = kuranas_fold($3) THEN 0
        WHEN kuranas_fold(hf.name) LIKE kuranas_fold($4) THEN 1
        WHEN kuranas_fold(hf.name) LIKE kuranas_fold($5) THEN 2
        ELSE 3
    END,
    hf.starred DESC,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $7;
