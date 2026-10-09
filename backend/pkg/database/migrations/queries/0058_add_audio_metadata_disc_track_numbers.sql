ALTER TABLE audio_metadata
    ADD COLUMN IF NOT EXISTS disc_number INTEGER NULL,
    ADD COLUMN IF NOT EXISTS disc_total INTEGER NULL,
    ADD COLUMN IF NOT EXISTS track_no INTEGER NULL,
    ADD COLUMN IF NOT EXISTS track_total INTEGER NULL,
    ADD COLUMN IF NOT EXISTS tags_extracted_version INTEGER NOT NULL DEFAULT 0;

UPDATE audio_metadata
SET
    track_no = NULLIF(substring(track_number FROM '^\s*(\d{1,9})'), '')::INTEGER,
    track_total = NULLIF(substring(track_number FROM '/\s*(\d{1,9})'), '')::INTEGER
WHERE track_number IS NOT NULL
  AND track_no IS NULL;
