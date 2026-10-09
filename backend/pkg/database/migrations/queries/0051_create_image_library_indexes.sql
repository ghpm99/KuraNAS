CREATE INDEX IF NOT EXISTS image_metadata_library_taken_at
    ON image_metadata ((COALESCE(taken_at, '-infinity'::TIMESTAMPTZ)) DESC, file_id DESC);
