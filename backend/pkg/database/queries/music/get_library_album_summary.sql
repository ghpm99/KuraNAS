SELECT
    am.catalog_album_key,
    MIN(am.catalog_album_label COLLATE "C") AS album_label,
    MIN(am.catalog_artist_label COLLATE "C") AS artist_label,
    COALESCE(
        (
            ARRAY_AGG(
                TRIM(am.year)
                ORDER BY
                    COALESCE(hf.updated_at, hf.created_at) DESC,
                    hf.id DESC
            ) FILTER (
                WHERE
                    TRIM(COALESCE(am.year, '')) <> ''
            )
        ) [1],
        ''
    ) AS year_label,
    COUNT(*) AS track_count,
    COALESCE(ROUND(SUM(am.length)), 0)::BIGINT AS total_length_seconds,
    COUNT(DISTINCT COALESCE(am.disc_number, 1)) AS disc_count
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.catalog_album_key = $2
    AND hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
GROUP BY
    am.catalog_album_key;
