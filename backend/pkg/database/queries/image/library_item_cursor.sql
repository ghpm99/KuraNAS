SELECT
    im.taken_at,
    im.file_id
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($1)
    AND im.file_id = $2
ORDER BY
    im.id DESC
LIMIT
    1
