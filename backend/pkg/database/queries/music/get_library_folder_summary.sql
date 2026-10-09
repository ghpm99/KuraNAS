SELECT
    COUNT(*) AS track_count,
    COALESCE(ROUND(SUM(am.length)), 0)::BIGINT AS total_length_seconds
FROM
    home_file hf
    LEFT JOIN audio_metadata am ON am.file_id = hf.id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND (
        hf.parent_path = $2
        OR hf.parent_path LIKE $3
    )
HAVING
    COUNT(*) > 0;
