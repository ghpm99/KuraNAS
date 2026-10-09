SELECT
    COUNT(*)
FROM
    player_queue pq
    JOIN home_file hf ON hf.id = pq.file_id
WHERE
    pq.client_id = $1
    AND hf.deleted_at IS NULL
    AND pq.position < COALESCE(
        (
            SELECT
                ps.queue_index
            FROM
                player_state ps
            WHERE
                ps.client_id = $1
        ),
        0
    );
