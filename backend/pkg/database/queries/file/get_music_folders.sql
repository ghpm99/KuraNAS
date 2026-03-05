SELECT
    hf.parent_path AS folder,
    COUNT(*) AS track_count
FROM
    home_file hf
    INNER JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
GROUP BY
    hf.parent_path
ORDER BY
    hf.parent_path
LIMIT
    $2
OFFSET
    $3;
