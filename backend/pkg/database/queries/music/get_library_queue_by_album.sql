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
    am.catalog_album_key = $2
    AND hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
ORDER BY
    COALESCE(am.disc_number, 1) ASC,
    am.track_no ASC NULLS LAST,
    COALESCE(NULLIF(TRIM(am.title), ''), hf."name") COLLATE "C" ASC,
    hf.id ASC
LIMIT
    $3;
