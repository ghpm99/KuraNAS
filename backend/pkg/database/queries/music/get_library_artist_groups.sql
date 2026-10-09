WITH artist_album_groups AS (
    SELECT
        am.catalog_artist_key AS artist_key,
        am.catalog_album_key AS album_key,
        MIN(am.catalog_artist_label COLLATE "C") AS artist_label,
        COUNT(*) AS track_count,
        MAX(hf.created_at) AS latest_added_at
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
    WHERE
        hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
        AND am.catalog_artist_key <> ''
    GROUP BY
        am.catalog_artist_key,
        am.catalog_album_key
),
artist_groups AS (
    SELECT
        artist_key,
        MIN(artist_label COLLATE "C") AS artist_label,
        SUM(track_count)::BIGINT AS track_count,
        COUNT(*) FILTER (
            WHERE
                album_key <> ''
        ) AS album_count,
        MAX(latest_added_at) AS latest_added_at
    FROM
        artist_album_groups
    GROUP BY
        artist_key
)
SELECT
    artist_key,
    artist_label,
    track_count,
    album_count
FROM
    artist_groups
ORDER BY
    (
        CASE
            WHEN $2::TEXT = 'tracks' THEN track_count
            WHEN $2::TEXT = 'recent' THEN EXTRACT(EPOCH FROM latest_added_at)::BIGINT
        END
    ) * (CASE WHEN $3::BOOLEAN THEN -1 ELSE 1 END) ASC NULLS LAST,
    (CASE WHEN $2::TEXT = 'name' AND NOT $3::BOOLEAN THEN artist_label END) COLLATE "C" ASC,
    (CASE WHEN $2::TEXT = 'name' AND $3::BOOLEAN THEN artist_label END) COLLATE "C" DESC,
    track_count DESC,
    artist_label COLLATE "C" ASC,
    artist_key ASC
LIMIT
    $4 OFFSET $5;
