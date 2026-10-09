INSERT INTO player_queue (client_id, position, file_id)
SELECT
    $1,
    queued.ordinal - 1,
    queued.file_id
FROM
    unnest($2::int[]) WITH ORDINALITY AS queued (file_id, ordinal)
    JOIN home_file hf ON hf.id = queued.file_id;
