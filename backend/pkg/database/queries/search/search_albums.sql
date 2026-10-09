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
        AND hf.format = ANY($5)
        AND TRIM(am.album) <> ''
        AND COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), '')) <> ''
        AND am.file_id IN (
            SELECT album_match.file_id
            FROM audio_metadata album_match
            WHERE lower(album_match.album) LIKE $1
                AND lower(album_match.album) LIKE ALL ($2::text[])
            UNION
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
            lower(TRIM(am.album)) LIKE ALL ($2::text[])
            OR lower(COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), ''))) LIKE ALL ($2::text[])
            OR lower(hf.name) LIKE ALL ($2::text[])
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
        WHEN lower(album) = $3 THEN 0
        WHEN lower(album) LIKE $4 THEN 1
        ELSE 2
    END,
    track_count DESC,
    artist ASC,
    album ASC
LIMIT
    $6;
