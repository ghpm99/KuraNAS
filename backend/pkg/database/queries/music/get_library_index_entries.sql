SELECT
    hf.id,
    hf."name",
    hf."path",
    hf.parent_path,
    hf.starred,
    hf.created_at,
    hf.updated_at,
    hf.last_interaction,
    COALESCE(NULLIF(TRIM(am.title), ''), hf."name") AS title,
    COALESCE(NULLIF(TRIM(am.artist), ''), '') AS artist,
    COALESCE(NULLIF(TRIM(am.album_artist), ''), '') AS album_artist,
    COALESCE(NULLIF(TRIM(am.album), ''), '') AS album,
    COALESCE(NULLIF(TRIM(am.genre), ''), '') AS genre,
    COALESCE(NULLIF(TRIM(am.YEAR), ''), '') AS year,
    COALESCE(NULLIF(TRIM(am.track_number), ''), '') AS track_number
FROM
    home_file hf
    LEFT JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
ORDER BY
    COALESCE(hf.last_interaction, hf.updated_at, hf.created_at) DESC,
    hf.id DESC;
