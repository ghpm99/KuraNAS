SELECT
    hf.id,
    COUNT(*) AS play_count,
    MAX(mpe.played_at) AS last_played_at
FROM
    music_play_event mpe
    JOIN home_file hf ON hf.id = mpe.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND ($2::TIMESTAMPTZ IS NULL OR mpe.played_at >= $2)
GROUP BY
    hf.id
ORDER BY
    play_count DESC,
    last_played_at DESC,
    hf.id DESC
LIMIT
    $3 OFFSET $4;
