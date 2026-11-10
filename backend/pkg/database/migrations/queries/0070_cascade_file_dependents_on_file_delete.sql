ALTER TABLE image_metadata
    DROP CONSTRAINT IF EXISTS image_metadata_file_id_fkey;

ALTER TABLE image_metadata
    ADD CONSTRAINT image_metadata_file_id_fkey
    FOREIGN KEY (file_id) REFERENCES home_file (id) ON DELETE CASCADE;

ALTER TABLE audio_metadata
    DROP CONSTRAINT IF EXISTS audio_metadata_file_id_fkey;

ALTER TABLE audio_metadata
    ADD CONSTRAINT audio_metadata_file_id_fkey
    FOREIGN KEY (file_id) REFERENCES home_file (id) ON DELETE CASCADE;

ALTER TABLE video_metadata
    DROP CONSTRAINT IF EXISTS video_metadata_file_id_fkey;

ALTER TABLE video_metadata
    ADD CONSTRAINT video_metadata_file_id_fkey
    FOREIGN KEY (file_id) REFERENCES home_file (id) ON DELETE CASCADE;

DELETE FROM recent_file
WHERE NOT EXISTS (SELECT 1 FROM home_file WHERE home_file.id = recent_file.file_id);

ALTER TABLE recent_file
    DROP CONSTRAINT IF EXISTS recent_file_file_id_fkey;

ALTER TABLE recent_file
    ADD CONSTRAINT recent_file_file_id_fkey
    FOREIGN KEY (file_id) REFERENCES home_file (id) ON DELETE CASCADE;
