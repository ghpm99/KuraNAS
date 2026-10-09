DELETE FROM player_queue
WHERE
    client_id = $1;
