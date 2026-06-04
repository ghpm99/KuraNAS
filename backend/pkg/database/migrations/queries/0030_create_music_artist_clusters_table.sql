ALTER TABLE playlist
    ADD COLUMN IF NOT EXISTS is_ai_generated BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS music_artist_clusters (
    artist_key   TEXT PRIMARY KEY,
    artist       TEXT NOT NULL,
    cluster_name TEXT NOT NULL,
    updated_at   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_music_artist_clusters_name
    ON music_artist_clusters (cluster_name);
