CREATE TABLE IF NOT EXISTS player_queue (
    client_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    file_id INTEGER NOT NULL REFERENCES home_file(id) ON DELETE CASCADE,
    PRIMARY KEY (client_id, position)
);

ALTER TABLE player_state
    ADD COLUMN IF NOT EXISTS queue_index INTEGER NOT NULL DEFAULT 0;
