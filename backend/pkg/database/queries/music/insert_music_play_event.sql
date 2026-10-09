INSERT INTO
    music_play_event (file_id, client_id, played_seconds)
SELECT
    hf.id,
    $3,
    $4
FROM
    home_file hf
WHERE
    hf.id = $1
    AND hf.format = ANY ($2)
    AND hf.deleted_at IS NULL;
