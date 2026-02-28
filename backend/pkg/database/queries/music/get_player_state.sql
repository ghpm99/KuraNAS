SELECT
    ps.id,
    ps.client_id,
    ps.playlist_id,
    ps.current_file_id,
    ps.current_position,
    ps.volume,
    ps.shuffle,
    ps.repeat_mode,
    ps.updated_at
FROM
    player_state ps
WHERE
    ps.client_id = $1;
