SELECT
    hf.id,
    hf."path"
FROM
    home_file hf
WHERE
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
    AND hf.id > $2
    AND NOT EXISTS (
        SELECT
            1
        FROM
            audio_metadata am
        WHERE
            am.file_id = hf.id
    )
ORDER BY
    hf.id
LIMIT
    $3;
