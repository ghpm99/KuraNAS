CREATE TABLE IF NOT EXISTS video_playlist (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('folder', 'series', 'movie', 'custom')),
    source_path TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_played_at TIMESTAMP,
    UNIQUE (type, source_path)
);

CREATE TABLE IF NOT EXISTS video_playlist_item (
    id SERIAL PRIMARY KEY,
    playlist_id INTEGER NOT NULL REFERENCES video_playlist(id) ON DELETE CASCADE,
    video_id INTEGER NOT NULL REFERENCES home_file(id) ON DELETE CASCADE,
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (playlist_id, video_id),
    UNIQUE (playlist_id, order_index)
);

CREATE TABLE IF NOT EXISTS video_playback_state (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(128) NOT NULL UNIQUE,
    playlist_id INTEGER REFERENCES video_playlist(id) ON DELETE SET NULL,
    video_id INTEGER REFERENCES home_file(id) ON DELETE SET NULL,
    current_position DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_paused BOOLEAN NOT NULL DEFAULT TRUE,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_update TIMESTAMP NOT NULL DEFAULT NOW()
);
