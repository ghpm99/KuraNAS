WITH ranked_albums AS (
    SELECT
        COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) AS artist,
        TRIM(am.album) AS album,
        MAX(TRIM(am.year)) AS year,
        COUNT(*) AS track_count
    FROM
        home_file hf
        INNER JOIN audio_metadata am ON hf.id = am.file_id
    WHERE
        hf.deleted_at IS NULL
        AND hf.format = ANY($2)
        AND TRIM(am.album) <> ''
        AND COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) <> ''
        AND (
            TRIM(am.album) ILIKE '%' || $1 || '%'
            OR COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) ILIKE '%' || $1 || '%'
            OR hf.name ILIKE '%' || $1 || '%'
            OR hf.path ILIKE '%' || $1 || '%'
        )
    GROUP BY
        COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')),
        TRIM(am.album)
)
SELECT
    artist,
    album,
    year,
    track_count
FROM
    ranked_albums
ORDER BY
    CASE
        WHEN LOWER(album) = LOWER($1) THEN 0
        WHEN album ILIKE $1 || '%' THEN 1
        ELSE 2
    END,
    track_count DESC,
    artist ASC,
    album ASC
LIMIT
    $3;
