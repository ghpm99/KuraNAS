SELECT
    hf.id,
    hf."name",
    hf."path",
    hf.format,
    COALESCE(am.title, ''),
    COALESCE(am.artist, ''),
    COALESCE(am.album, ''),
    COALESCE(am.LENGTH, 0)
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.catalog_genre_keys @> ARRAY[$2]::TEXT[]
    AND hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
ORDER BY
    am.catalog_artist_label COLLATE "C" ASC,
    am.catalog_album_label COLLATE "C" ASC,
    COALESCE(am.disc_number, 1) ASC,
    am.track_no ASC NULLS LAST,
    COALESCE(NULLIF(TRIM(am.title), ''), hf."name") COLLATE "C" ASC,
    hf.id ASC
LIMIT
    $3;
