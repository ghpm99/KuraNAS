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
            WHERE kuranas_fold(album_match.album) LIKE kuranas_fold($1)
                AND kuranas_fold(album_match.album) LIKE ALL (kuranas_fold_terms($2::text[]))
            UNION
            SELECT album_artist_match.file_id
            FROM audio_metadata album_artist_match
            WHERE kuranas_fold(album_artist_match.album_artist) LIKE kuranas_fold($1)
                AND kuranas_fold(album_artist_match.album_artist) LIKE ALL (kuranas_fold_terms($2::text[]))
            UNION
            SELECT artist_match.file_id
            FROM audio_metadata artist_match
            WHERE kuranas_fold(artist_match.artist) LIKE kuranas_fold($1)
                AND kuranas_fold(artist_match.artist) LIKE ALL (kuranas_fold_terms($2::text[]))
            UNION
            SELECT name_match.id
            FROM home_file name_match
            WHERE name_match.deleted_at IS NULL
                AND kuranas_fold(name_match.name) LIKE kuranas_fold($1)
                AND kuranas_fold(name_match.name) LIKE ALL (kuranas_fold_terms($2::text[]))
        )
        AND (
            kuranas_fold(TRIM(am.album)) LIKE ALL (kuranas_fold_terms($2::text[]))
            OR kuranas_fold(COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), ''))) LIKE ALL (kuranas_fold_terms($2::text[]))
            OR kuranas_fold(hf.name) LIKE ALL (kuranas_fold_terms($2::text[]))
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
        WHEN kuranas_fold(album) = kuranas_fold($3) THEN 0
        WHEN kuranas_fold(album) LIKE kuranas_fold($4) THEN 1
        ELSE 2
    END,
    track_count DESC,
    artist ASC,
    album ASC
LIMIT
    $6;
