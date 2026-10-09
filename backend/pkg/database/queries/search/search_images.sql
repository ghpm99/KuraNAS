SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
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
            AND lower(name_match.name) LIKE $1
            AND lower(name_match.name) LIKE ALL ($2::text[])
        UNION
        SELECT content_match.file_id
        FROM image_metadata content_match
        WHERE content_match.ai_search_text LIKE $1
            AND content_match.ai_search_text LIKE ALL ($2::text[])
    )
ORDER BY
    CASE
        WHEN lower(hf.name) = $3 THEN 0
        WHEN lower(hf.name) LIKE $4 THEN 1
        WHEN lower(hf.name) LIKE $5 THEN 2
        ELSE 3
    END,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $7;
