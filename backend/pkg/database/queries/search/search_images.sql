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
    AND hf.format = ANY($2)
    AND (
        hf.name ILIKE '%' || $1 || '%'
        OR hf.path ILIKE '%' || $1 || '%'
        OR COALESCE(im.make, '') ILIKE '%' || $1 || '%'
        OR COALESCE(im.model, '') ILIKE '%' || $1 || '%'
        OR COALESCE(im.artist, '') ILIKE '%' || $1 || '%'
        OR COALESCE(im.image_description, '') ILIKE '%' || $1 || '%'
    )
ORDER BY
    CASE
        WHEN LOWER(hf.name) = LOWER($1) THEN 0
        WHEN hf.name ILIKE $1 || '%' THEN 1
        ELSE 2
    END,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $3;
