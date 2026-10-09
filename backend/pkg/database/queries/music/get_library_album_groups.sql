WITH album_groups AS (
    SELECT
        am.catalog_album_key AS album_key,
        MIN(am.catalog_album_label COLLATE "C") AS album_label,
        MIN(am.catalog_artist_label COLLATE "C") AS artist_label,
        COUNT(*) AS track_count
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
    WHERE
        hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
        AND am.catalog_album_key <> ''
    GROUP BY
        am.catalog_album_key
),
album_page AS (
    SELECT
        album_key,
        album_label,
        artist_label,
        track_count
    FROM
        album_groups
    ORDER BY
        track_count DESC,
        artist_label COLLATE "C" ASC,
        album_label COLLATE "C" ASC,
        album_key ASC
    LIMIT
        $2 OFFSET $3
)
SELECT
    album_page.album_key,
    album_page.album_label,
    album_page.artist_label,
    COALESCE(album_year.year_label, '') AS year_label,
    album_page.track_count
FROM
    album_page
    LEFT JOIN LATERAL (
        SELECT
            TRIM(am.year) AS year_label
        FROM
            audio_metadata am
            JOIN home_file hf ON hf.id = am.file_id
        WHERE
            am.catalog_album_key = album_page.album_key
            AND hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
            AND TRIM(COALESCE(am.year, '')) <> ''
        ORDER BY
            COALESCE(hf.updated_at, hf.created_at) DESC,
            hf.id DESC
        LIMIT
            1
    ) album_year ON TRUE
ORDER BY
    album_page.track_count DESC,
    album_page.artist_label COLLATE "C" ASC,
    album_page.album_label COLLATE "C" ASC,
    album_page.album_key ASC;
