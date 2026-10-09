INSERT INTO player_state (client_id, queue_index, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (client_id)
DO UPDATE SET
    queue_index = EXCLUDED.queue_index,
    updated_at = CURRENT_TIMESTAMP;
