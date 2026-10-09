SELECT
    hf.id,
    hf."path"
FROM
    home_file hf
WHERE
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
    AND hf.id > $2
    AND EXISTS (
        SELECT
            1
        FROM
            audio_metadata am
        WHERE
            am.file_id = hf.id
            AND am.tags_extracted_version < $3
            AND (
                (
                    COALESCE(am.artist, '') = ''
                    AND COALESCE(am.album, '') = ''
                    AND COALESCE(am.title, '') = ''
                )
                OR COALESCE(am."year", '') = ''
            )
    )
ORDER BY
    hf.id
LIMIT
    $4;
