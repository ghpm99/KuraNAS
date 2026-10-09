SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.updated_at,
    COALESCE(im.classification_category, ''),
    COALESCE(NULLIF(TRIM(im.model), ''), NULLIF(TRIM(im.make), ''), NULLIF(TRIM(im.artist), ''), '')
FROM
    home_file hf
    LEFT JOIN image_metadata im ON hf.id = im.file_id
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($6)
    AND hf.id IN (
        SELECT name_match.id
        FROM home_file name_match
        WHERE name_match.deleted_at IS NULL
            AND kuranas_fold(name_match.name) LIKE kuranas_fold($1)
            AND kuranas_fold(name_match.name) LIKE ALL (kuranas_fold_terms($2::text[]))
        UNION
        SELECT content_match.file_id
        FROM image_metadata content_match
        WHERE kuranas_fold(content_match.ai_search_text) LIKE kuranas_fold($1)
            AND kuranas_fold(content_match.ai_search_text) LIKE ALL (kuranas_fold_terms($2::text[]))
    )
ORDER BY
    CASE
        WHEN kuranas_fold(hf.name) = kuranas_fold($3) THEN 0
        WHEN kuranas_fold(hf.name) LIKE kuranas_fold($4) THEN 1
        WHEN kuranas_fold(hf.name) LIKE kuranas_fold($5) THEN 2
        ELSE 3
    END,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $7;
