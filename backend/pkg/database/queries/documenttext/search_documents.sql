SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.updated_at,
    dt.extracted_text
FROM
    document_text dt
    JOIN home_file hf ON hf.id = dt.file_id
WHERE
    hf.deleted_at IS NULL
    AND dt.error IS NULL
    AND kuranas_fold(dt.extracted_text) LIKE kuranas_fold($1)
    AND kuranas_fold(dt.extracted_text) LIKE ALL (kuranas_fold_terms($2::text[]))
ORDER BY
    CASE
        WHEN kuranas_fold(hf.name) LIKE ALL (kuranas_fold_terms($2::text[])) THEN 0
        ELSE 1
    END,
    hf.updated_at DESC,
    hf.id ASC
LIMIT
    $3 OFFSET $4;
