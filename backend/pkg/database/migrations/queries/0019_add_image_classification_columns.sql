ALTER TABLE image_metadata
ADD COLUMN IF NOT EXISTS classification_category TEXT NOT NULL DEFAULT 'other',
ADD COLUMN IF NOT EXISTS classification_confidence REAL NOT NULL DEFAULT 0;
