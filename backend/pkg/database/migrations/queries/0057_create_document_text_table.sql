CREATE TABLE IF NOT EXISTS document_text (
    file_id INTEGER PRIMARY KEY REFERENCES home_file (id) ON DELETE CASCADE,
    extracted_text TEXT NOT NULL DEFAULT '',
    text_length INTEGER NOT NULL DEFAULT 0,
    truncated BOOLEAN NOT NULL DEFAULT FALSE,
    extracted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source_updated_at TIMESTAMPTZ NOT NULL,
    error TEXT NULL
);

DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'pg_trgm extension unavailable, document text search stays unindexed: %', SQLERRM;
    END;

    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        RETURN;
    END IF;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_document_text_extracted_text_fold_trigram
            ON document_text USING gin (kuranas_fold(extracted_text) gin_trgm_ops);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'document text fold trigram index not created: %', SQLERRM;
    END;
END
$$;
