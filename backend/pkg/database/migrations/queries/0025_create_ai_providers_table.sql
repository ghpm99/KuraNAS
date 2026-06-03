CREATE TABLE IF NOT EXISTS ai_providers (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(40) NOT NULL UNIQUE,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    model       VARCHAR(200) NOT NULL DEFAULT '',
    base_url    VARCHAR(500) NOT NULL DEFAULT '',
    priority    INTEGER NOT NULL DEFAULT 0,
    params      JSON DEFAULT '{}',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ai_providers_name_check CHECK (
        name IN ('ollama', 'openai', 'anthropic')
    )
);
