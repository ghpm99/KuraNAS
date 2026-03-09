CREATE TABLE IF NOT EXISTS video_behavior_event (
    id          SERIAL PRIMARY KEY,
    client_id   VARCHAR(128) NOT NULL,
    video_id    INTEGER REFERENCES home_file(id) ON DELETE CASCADE,
    playlist_id INTEGER REFERENCES video_playlist(id) ON DELETE SET NULL,
    event_type  VARCHAR(20) NOT NULL,
    position    DOUBLE PRECISION DEFAULT 0,
    duration    DOUBLE PRECISION DEFAULT 0,
    watched_pct DOUBLE PRECISION DEFAULT 0,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_behavior_event_client
    ON video_behavior_event(client_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_behavior_event_video
    ON video_behavior_event(video_id);
