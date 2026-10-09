SELECT
    am.catalog_artist_key,
    MIN(am.catalog_artist_label COLLATE "C") AS artist_label,
    COUNT(*) AS track_count,
    COUNT(DISTINCT am.catalog_album_key) FILTER (
        WHERE
            am.catalog_album_key <> ''
    ) AS album_count,
    COALESCE(ROUND(SUM(am.length)), 0)::BIGINT AS total_length_seconds
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.catalog_artist_key = $2
    AND hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
GROUP BY
    am.catalog_artist_key;
