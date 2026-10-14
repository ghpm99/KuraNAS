ALTER TABLE video_playlist
    ADD COLUMN IF NOT EXISTS name_key VARCHAR(80);
