SELECT
    hf.id
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
    COALESCE(am.track_no, 0) ASC,
    COALESCE(NULLIF(TRIM(am.title), ''), hf."name") COLLATE "C" ASC,
    hf.id ASC
LIMIT
    $3 OFFSET $4;
