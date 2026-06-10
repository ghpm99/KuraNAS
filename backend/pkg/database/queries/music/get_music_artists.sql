SELECT
    am.artist,
    COUNT(*) AS track_count,
    COUNT(DISTINCT am.album) AS album_count
FROM
    home_file hf
    INNER JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND am.artist != ''
GROUP BY
    am.artist
ORDER BY
    am.artist
LIMIT
    $2
OFFSET
    $3;
