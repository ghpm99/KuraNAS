SELECT
    hf.id,
    hf."name",
    hf."path",
    hf.format,
    COALESCE(am.title, ''),
    COALESCE(am.artist, ''),
    COALESCE(am.album, ''),
    COALESCE(am.LENGTH, 0)
FROM
    home_file hf
    LEFT JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND hf.id = ANY ($2)
ORDER BY
    array_position($2::int[], hf.id);
