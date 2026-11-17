SELECT
    am.id,
    am.file_id
FROM
    audio_metadata am
WHERE
    (
        am.catalog_artist_key IS NULL
        OR am.catalog_keys_version < $3
    )
    AND am.id > $1
ORDER BY
    am.id
LIMIT
    $2;
