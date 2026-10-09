SELECT
    COUNT(*)
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
WHERE
    (
        (im.ai_classified_at IS NULL AND im.classification_confidence < $1)
        OR (im.ai_classified_at IS NOT NULL AND im.ai_caption IS NULL)
    )
    AND hf.deleted_at IS NULL;
