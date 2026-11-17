SELECT DISTINCT
    hf.id
FROM
    home_file hf
    JOIN audio_metadata am ON am.file_id = hf.id
WHERE
    hf.id = ANY ($1)
    AND hf.deleted_at IS NULL
ORDER BY
    hf.id ASC;
