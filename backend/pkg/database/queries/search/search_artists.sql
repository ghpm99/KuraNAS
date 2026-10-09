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
        AND hf.format = ANY($5)
        AND COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) <> ''
        AND am.file_id IN (
            SELECT album_artist_match.file_id
            FROM audio_metadata album_artist_match
            WHERE lower(album_artist_match.album_artist) LIKE $1
                AND lower(album_artist_match.album_artist) LIKE ALL ($2::text[])
            UNION
            SELECT artist_match.file_id
            FROM audio_metadata artist_match
            WHERE lower(artist_match.artist) LIKE $1
                AND lower(artist_match.artist) LIKE ALL ($2::text[])
            UNION
            SELECT name_match.id
            FROM home_file name_match
            WHERE name_match.deleted_at IS NULL
                AND lower(name_match.name) LIKE $1
                AND lower(name_match.name) LIKE ALL ($2::text[])
        )
        AND (
            lower(COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), ''))) LIKE ALL ($2::text[])
            OR lower(hf.name) LIKE ALL ($2::text[])
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
        WHEN lower(artist) = $3 THEN 0
        WHEN lower(artist) LIKE $4 THEN 1
        ELSE 2
    END,
    track_count DESC,
    artist ASC
LIMIT
    $6;
