WITH genre_groups AS (
    SELECT
        genre.genre_key,
        MIN(genre.genre_label COLLATE "C") AS genre_label,
        COUNT(*) AS track_count
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
        CROSS JOIN LATERAL unnest(am.catalog_genre_keys, am.catalog_genre_labels) AS genre (genre_key, genre_label)
    WHERE
        hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
    GROUP BY
        genre.genre_key
)
SELECT
    genre_key,
    genre_label,
    track_count
FROM
    genre_groups
ORDER BY
    track_count DESC,
    genre_label COLLATE "C" ASC,
    genre_key ASC
LIMIT
    $2 OFFSET $3;
