INSERT INTO player_state (client_id, playlist_id, current_file_id, current_position, volume, shuffle, repeat_mode, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
ON CONFLICT (client_id)
DO UPDATE SET
    playlist_id = EXCLUDED.playlist_id,
    current_file_id = EXCLUDED.current_file_id,
    current_position = EXCLUDED.current_position,
    volume = EXCLUDED.volume,
    shuffle = EXCLUDED.shuffle,
    repeat_mode = EXCLUDED.repeat_mode,
    updated_at = CURRENT_TIMESTAMP
RETURNING id, updated_at;
