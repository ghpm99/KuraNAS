ALTER TABLE image_metadata
ADD COLUMN IF NOT EXISTS ai_classified_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_image_metadata_pending_ai_classification
    ON image_metadata (file_id)
    WHERE ai_classified_at IS NULL;
