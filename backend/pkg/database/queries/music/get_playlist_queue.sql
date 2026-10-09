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
    playlist_track pt
    INNER JOIN home_file hf ON pt.file_id = hf.id
    LEFT JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    pt.playlist_id = $1
    AND hf.deleted_at IS NULL
ORDER BY
    pt.position
LIMIT
    $2;
