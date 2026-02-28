CREATE TABLE IF NOT EXISTS player_state (
    id SERIAL PRIMARY KEY,
    client_id TEXT NOT NULL UNIQUE,
    playlist_id INTEGER REFERENCES playlist(id) ON DELETE SET NULL,
    current_file_id INTEGER REFERENCES home_file(id) ON DELETE SET NULL,
    current_position REAL DEFAULT 0,
    volume REAL DEFAULT 1.0,
    shuffle BOOLEAN DEFAULT FALSE,
    repeat_mode TEXT DEFAULT 'none',
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
