WITH genre_groups AS (
    SELECT
        genre.genre_key,
        MIN(genre.genre_label COLLATE "C") AS genre_label,
        COUNT(*) AS track_count,
        MAX(hf.created_at) AS latest_added_at
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
    (
        CASE
            WHEN $2::TEXT = 'tracks' THEN track_count
            WHEN $2::TEXT = 'recent' THEN EXTRACT(EPOCH FROM latest_added_at)::BIGINT
        END
    ) * (CASE WHEN $3::BOOLEAN THEN -1 ELSE 1 END) ASC NULLS LAST,
    (CASE WHEN $2::TEXT = 'name' AND NOT $3::BOOLEAN THEN genre_label END) COLLATE "C" ASC,
    (CASE WHEN $2::TEXT = 'name' AND $3::BOOLEAN THEN genre_label END) COLLATE "C" DESC,
    track_count DESC,
    genre_label COLLATE "C" ASC,
    genre_key ASC
LIMIT
    $4 OFFSET $5;
