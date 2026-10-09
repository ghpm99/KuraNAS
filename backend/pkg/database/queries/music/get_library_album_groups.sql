WITH album_groups AS (
    SELECT
        am.catalog_album_key AS album_key,
        MIN(am.catalog_album_label COLLATE "C") AS album_label,
        MIN(am.catalog_artist_label COLLATE "C") AS artist_label,
        COUNT(*) AS track_count,
        MAX(hf.created_at) AS latest_added_at,
        MAX(NULLIF(substring(TRIM(COALESCE(am.year, '')) FROM '^\d{4}'), '')::INT) AS sort_year
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
        track_count,
        ROW_NUMBER() OVER (
            ORDER BY
                (
                    CASE
                        WHEN $2::TEXT = 'tracks' THEN track_count
                        WHEN $2::TEXT = 'recent' THEN EXTRACT(EPOCH FROM latest_added_at)::BIGINT
                        WHEN $2::TEXT = 'year' THEN sort_year
                    END
                ) * (CASE WHEN $3::BOOLEAN THEN -1 ELSE 1 END) ASC NULLS LAST,
                (CASE WHEN $2::TEXT = 'name' AND NOT $3::BOOLEAN THEN album_label END) COLLATE "C" ASC,
                (CASE WHEN $2::TEXT = 'name' AND $3::BOOLEAN THEN album_label END) COLLATE "C" DESC,
                track_count DESC,
                artist_label COLLATE "C" ASC,
                album_label COLLATE "C" ASC,
                album_key ASC
        ) AS sort_position
    FROM
        album_groups
    ORDER BY
        sort_position ASC
    LIMIT
        $4 OFFSET $5
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
    album_page.sort_position ASC;
