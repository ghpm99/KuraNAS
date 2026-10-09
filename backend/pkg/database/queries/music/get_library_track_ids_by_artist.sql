SELECT
    hf.id
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.catalog_artist_key = $2
    AND hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
ORDER BY
    MIN(NULLIF(substring(TRIM(COALESCE(am.year, '')) FROM '^\d{4}'), '')) OVER (PARTITION BY am.catalog_album_key) ASC NULLS LAST,
    am.catalog_album_label COLLATE "C" ASC,
    am.catalog_album_key ASC,
    COALESCE(am.disc_number, 1) ASC,
    am.track_no ASC NULLS LAST,
    COALESCE(NULLIF(TRIM(am.title), ''), hf."name") COLLATE "C" ASC,
    hf.id ASC
LIMIT
    $3 OFFSET $4;
