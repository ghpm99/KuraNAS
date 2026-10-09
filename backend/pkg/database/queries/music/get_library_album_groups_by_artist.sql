WITH artist_albums AS (
    SELECT
        am.catalog_album_key AS album_key,
        MIN(am.catalog_album_label COLLATE "C") AS album_label,
        COALESCE(
        MIN(am.catalog_album_artist_label COLLATE "C"),
        MIN(am.catalog_artist_label COLLATE "C")
    ) AS artist_label,
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
        MAX(NULLIF(substring(TRIM(COALESCE(am.year, '')) FROM '^\d{4}'), '')::INT) AS sort_year
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
    WHERE
        am.catalog_artist_key = $2
        AND am.catalog_album_key <> ''
        AND hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
    GROUP BY
        am.catalog_album_key
)
SELECT
    album_key,
    album_label,
    artist_label,
    year_label,
    track_count
FROM
    artist_albums
ORDER BY
    sort_year ASC NULLS LAST,
    album_label COLLATE "C" ASC,
    album_key ASC
LIMIT
    $3 OFFSET $4;
