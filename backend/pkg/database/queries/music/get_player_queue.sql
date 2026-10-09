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
    player_queue pq
    JOIN home_file hf ON hf.id = pq.file_id
    LEFT JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    pq.client_id = $1
    AND hf.deleted_at IS NULL
ORDER BY
    pq.position;
