SELECT
    genre.genre_key,
    MIN(genre.genre_label COLLATE "C") AS genre_label,
    COUNT(*) AS track_count,
    COALESCE(ROUND(SUM(am.length)), 0)::BIGINT AS total_length_seconds
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
    CROSS JOIN LATERAL unnest(am.catalog_genre_keys, am.catalog_genre_labels) AS genre (genre_key, genre_label)
WHERE
    genre.genre_key = $2
    AND hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
GROUP BY
    genre.genre_key;
