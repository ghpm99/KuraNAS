-- AI vision: a descriptive filename suggested by the model from the image
-- content, so users can recognize files like "55.jpg" or "untitled(1).jpeg".
ALTER TABLE image_metadata
    ADD COLUMN IF NOT EXISTS classification_suggested_name TEXT NOT NULL DEFAULT '';
