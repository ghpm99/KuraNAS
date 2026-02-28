CREATE TABLE IF NOT EXISTS playlist_track (
    id SERIAL PRIMARY KEY,
    playlist_id INTEGER NOT NULL REFERENCES playlist(id) ON DELETE CASCADE,
    file_id INTEGER NOT NULL REFERENCES home_file(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (playlist_id, file_id)
);

CREATE INDEX IF NOT EXISTS idx_playlist_track_position ON playlist_track(playlist_id, position);
