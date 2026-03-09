ALTER TABLE video_playlist
    DROP CONSTRAINT IF EXISTS video_playlist_type_check;

ALTER TABLE video_playlist
    ADD CONSTRAINT video_playlist_type_check
    CHECK (type IN ('folder', 'series', 'movie', 'custom', 'course', 'mixed', 'continue'));
