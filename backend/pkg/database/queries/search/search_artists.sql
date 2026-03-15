WITH ranked_artists AS (
    SELECT
        COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) AS artist,
        COUNT(*) AS track_count,
        COUNT(DISTINCT NULLIF(TRIM(am.album), '')) AS album_count
    FROM
        home_file hf
        INNER JOIN audio_metadata am ON hf.id = am.file_id
    WHERE
        hf.deleted_at IS NULL
        AND hf.format = ANY($2)
        AND COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) <> ''
        AND (
            COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) ILIKE '%' || $1 || '%'
            OR hf.name ILIKE '%' || $1 || '%'
            OR hf.path ILIKE '%' || $1 || '%'
        )
    GROUP BY
        COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), ''))
)
SELECT
    artist,
    track_count,
    album_count
FROM
    ranked_artists
ORDER BY
    CASE
        WHEN LOWER(artist) = LOWER($1) THEN 0
        WHEN artist ILIKE $1 || '%' THEN 1
        ELSE 2
    END,
    track_count DESC,
    artist ASC
LIMIT
    $3;
