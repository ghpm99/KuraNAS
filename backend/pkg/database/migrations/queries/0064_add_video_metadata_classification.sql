ALTER TABLE video_metadata
    ADD COLUMN IF NOT EXISTS classification TEXT NULL,
    ADD COLUMN IF NOT EXISTS classification_version INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_video_metadata_classification
    ON video_metadata (classification);
