SELECT
    hf.id,
    hf."path"
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
WHERE
    im.ai_classified_at IS NULL
    AND im.classification_confidence < $1
    AND hf.deleted_at IS NULL
    AND hf.id > $2
ORDER BY
    hf.id
LIMIT
    $3;
