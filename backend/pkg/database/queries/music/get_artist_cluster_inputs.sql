WITH artist_totals AS (
    SELECT
        am.catalog_artist_key AS artist_key,
        MIN(am.catalog_artist_label COLLATE "C") AS artist_label,
        COUNT(*) AS track_count
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
    WHERE
        hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
        AND am.catalog_artist_key <> ''
    GROUP BY
        am.catalog_artist_key
),
artist_genre_counts AS (
    SELECT
        am.catalog_artist_key AS artist_key,
        genre.genre_label,
        COUNT(*) AS genre_track_count
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
        CROSS JOIN LATERAL unnest(am.catalog_genre_labels) AS genre (genre_label)
    WHERE
        hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
        AND am.catalog_artist_key <> ''
    GROUP BY
        am.catalog_artist_key,
        genre.genre_label
),
artist_top_genres AS (
    SELECT DISTINCT
        ON (artist_key) artist_key,
        genre_label
    FROM
        artist_genre_counts
    ORDER BY
        artist_key,
        genre_track_count DESC,
        genre_label COLLATE "C" ASC
)
SELECT
    artist_totals.artist_key,
    artist_totals.artist_label,
    COALESCE(artist_top_genres.genre_label, '') AS genre_hint,
    artist_totals.track_count
FROM
    artist_totals
    LEFT JOIN artist_top_genres ON artist_top_genres.artist_key = artist_totals.artist_key
ORDER BY
    artist_totals.track_count DESC,
    artist_totals.artist_label COLLATE "C" ASC,
    artist_totals.artist_key ASC;
