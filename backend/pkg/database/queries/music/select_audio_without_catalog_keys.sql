SELECT
    am.id,
    am.file_id
FROM
    audio_metadata am
WHERE
    am.catalog_artist_key IS NULL
    AND am.id > $1
ORDER BY
    am.id
LIMIT
    $2;
