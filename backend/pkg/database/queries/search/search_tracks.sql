SELECT
    hf.id,
    COALESCE(NULLIF(TRIM(am.title), ''), hf.name) AS title,
    COALESCE(NULLIF(TRIM(am.artist), ''), NULLIF(TRIM(am.album_artist), ''), '') AS artist,
    COALESCE(TRIM(am.album), '') AS album,
    COALESCE(NULLIF(TRIM(am.album_artist), ''), NULLIF(TRIM(am.artist), ''), '') AS album_owner,
    COALESCE(am.length, 0) AS duration,
    hf.path
FROM
    home_file hf
    LEFT JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($6)
    AND hf.id IN (
        SELECT title_match.file_id
        FROM audio_metadata title_match
        WHERE kuranas_fold(title_match.title) LIKE kuranas_fold($1)
        UNION
        SELECT artist_match.file_id
        FROM audio_metadata artist_match
        WHERE kuranas_fold(artist_match.artist) LIKE kuranas_fold($1)
        UNION
        SELECT album_match.file_id
        FROM audio_metadata album_match
        WHERE kuranas_fold(album_match.album) LIKE kuranas_fold($1)
        UNION
        SELECT name_match.id
        FROM home_file name_match
        WHERE name_match.deleted_at IS NULL
            AND kuranas_fold(name_match.name) LIKE kuranas_fold($1)
    )
    AND kuranas_fold(concat_ws(' ', am.title, am.artist, am.album, hf.name)) LIKE ALL (kuranas_fold_terms($2::text[]))
ORDER BY
    CASE
        WHEN kuranas_fold(COALESCE(NULLIF(TRIM(am.title), ''), hf.name)) = kuranas_fold($3) THEN 0
        WHEN kuranas_fold(COALESCE(NULLIF(TRIM(am.title), ''), hf.name)) LIKE kuranas_fold($4) THEN 1
        WHEN kuranas_fold(COALESCE(NULLIF(TRIM(am.title), ''), hf.name)) LIKE kuranas_fold($5) THEN 2
        ELSE 3
    END,
    hf.starred DESC,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $7;
