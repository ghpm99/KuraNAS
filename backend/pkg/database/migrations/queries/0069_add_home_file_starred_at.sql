ALTER TABLE home_file
    ADD COLUMN IF NOT EXISTS starred_at TIMESTAMPTZ NULL;

UPDATE home_file
SET starred_at = COALESCE(last_interaction, updated_at, created_at, now())
WHERE starred AND starred_at IS NULL;

CREATE INDEX IF NOT EXISTS home_file_active_starred_by_starred_at
    ON home_file (starred_at DESC, id DESC)
    WHERE starred AND deleted_at IS NULL;
