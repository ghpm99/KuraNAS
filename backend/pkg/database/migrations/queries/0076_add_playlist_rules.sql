ALTER TABLE playlist
    ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'manual';

ALTER TABLE playlist
    ADD COLUMN IF NOT EXISTS rules JSONB NULL;

UPDATE playlist
SET kind = CASE
    WHEN is_ai_generated THEN 'ai'
    WHEN is_system THEN 'system'
    ELSE 'manual'
END
WHERE kind = 'manual';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'playlist_kind_check'
    ) THEN
        ALTER TABLE playlist
            ADD CONSTRAINT playlist_kind_check CHECK (kind IN ('manual', 'rule', 'system', 'ai'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS playlist_rule_exclusion (
    playlist_id INTEGER NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    file_id INTEGER NOT NULL REFERENCES home_file (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (playlist_id, file_id)
);
