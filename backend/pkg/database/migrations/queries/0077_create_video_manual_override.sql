CREATE TABLE IF NOT EXISTS video_manual_override (
    file_id INTEGER PRIMARY KEY REFERENCES home_file (id) ON DELETE CASCADE,
    classification TEXT NULL CHECK (classification IN ('series', 'movie', 'personal', 'clip', 'other')),
    season INTEGER NULL CHECK (season >= 0),
    episode INTEGER NULL CHECK (episode >= 1),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE video_playlist_item
    ADD COLUMN IF NOT EXISTS season INTEGER NULL;

ALTER TABLE video_playlist_item
    ADD COLUMN IF NOT EXISTS episode INTEGER NULL;
