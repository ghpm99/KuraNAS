SELECT
    am.album,
    am.artist,
    am.YEAR,
    COUNT(*) AS track_count
FROM
    home_file hf
    INNER JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND am.album != ''
GROUP BY
    am.album,
    am.artist,
    am.YEAR
ORDER BY
    am.album
LIMIT
    $2
OFFSET
    $3;
