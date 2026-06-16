ALTER TABLE captures ADD COLUMN IF NOT EXISTS episode_key VARCHAR(512) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_captures_episode_key ON captures (episode_key) WHERE episode_key <> '';
