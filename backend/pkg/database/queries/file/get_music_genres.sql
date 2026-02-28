SELECT
    am.genre,
    COUNT(*) AS track_count
FROM
    home_file hf
    INNER JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND am.genre != ''
GROUP BY
    am.genre
ORDER BY
    am.genre
LIMIT
    $2
OFFSET
    $3;
