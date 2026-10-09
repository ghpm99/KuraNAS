CREATE TABLE IF NOT EXISTS music_play_event (
    id BIGSERIAL PRIMARY KEY,
    file_id INTEGER NOT NULL REFERENCES home_file(id) ON DELETE CASCADE,
    client_id TEXT NOT NULL,
    played_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    played_seconds INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_music_play_event_file_id ON music_play_event (file_id);

CREATE INDEX IF NOT EXISTS idx_music_play_event_played_at ON music_play_event (played_at DESC);
