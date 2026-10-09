SELECT
    am.id,
    COALESCE(am.artist, ''),
    COALESCE(am.album_artist, ''),
    COALESCE(am.album, ''),
    COALESCE(am.genre, '')
FROM
    audio_metadata am
WHERE
    am.catalog_artist_key IS NULL
    AND am.id > $1
ORDER BY
    am.id
LIMIT
    $2;
