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
        WHERE lower(title_match.title) LIKE $1
        UNION
        SELECT artist_match.file_id
        FROM audio_metadata artist_match
        WHERE lower(artist_match.artist) LIKE $1
        UNION
        SELECT album_match.file_id
        FROM audio_metadata album_match
        WHERE lower(album_match.album) LIKE $1
        UNION
        SELECT name_match.id
        FROM home_file name_match
        WHERE name_match.deleted_at IS NULL
            AND lower(name_match.name) LIKE $1
    )
    AND lower(concat_ws(' ', am.title, am.artist, am.album, hf.name)) LIKE ALL ($2::text[])
ORDER BY
    CASE
        WHEN lower(COALESCE(NULLIF(TRIM(am.title), ''), hf.name)) = $3 THEN 0
        WHEN lower(COALESCE(NULLIF(TRIM(am.title), ''), hf.name)) LIKE $4 THEN 1
        WHEN lower(COALESCE(NULLIF(TRIM(am.title), ''), hf.name)) LIKE $5 THEN 2
        ELSE 3
    END,
    hf.starred DESC,
    hf.updated_at DESC,
    hf.name ASC
LIMIT
    $7;
