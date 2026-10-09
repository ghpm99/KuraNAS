ALTER TABLE audio_metadata
    ADD COLUMN IF NOT EXISTS catalog_artist_key TEXT NULL,
    ADD COLUMN IF NOT EXISTS catalog_artist_label TEXT NULL,
    ADD COLUMN IF NOT EXISTS catalog_album_key TEXT NULL,
    ADD COLUMN IF NOT EXISTS catalog_album_label TEXT NULL,
    ADD COLUMN IF NOT EXISTS catalog_genre_keys TEXT[] NULL,
    ADD COLUMN IF NOT EXISTS catalog_genre_labels TEXT[] NULL;

CREATE INDEX IF NOT EXISTS audio_metadata_catalog_artist_key
    ON audio_metadata (catalog_artist_key, catalog_album_key, file_id)
    WHERE catalog_artist_key <> '';

CREATE INDEX IF NOT EXISTS audio_metadata_catalog_album_key
    ON audio_metadata (catalog_album_key, file_id)
    WHERE catalog_album_key <> '';

CREATE INDEX IF NOT EXISTS audio_metadata_catalog_genre_keys
    ON audio_metadata USING gin (catalog_genre_keys);

CREATE INDEX IF NOT EXISTS audio_metadata_catalog_keys_pending
    ON audio_metadata (id)
    WHERE catalog_artist_key IS NULL;

CREATE INDEX IF NOT EXISTS home_file_active_parent_path_pattern
    ON home_file (parent_path text_pattern_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS home_file_active_recent_activity
    ON home_file ((COALESCE(updated_at, created_at)) DESC, id DESC)
    WHERE deleted_at IS NULL;
