ALTER TABLE audio_metadata
    ADD COLUMN IF NOT EXISTS catalog_album_base_key TEXT NULL,
    ADD COLUMN IF NOT EXISTS catalog_album_artist_label TEXT NULL;

UPDATE audio_metadata
SET catalog_album_base_key = catalog_album_key
WHERE catalog_album_base_key IS NULL
    AND catalog_album_key IS NOT NULL;
