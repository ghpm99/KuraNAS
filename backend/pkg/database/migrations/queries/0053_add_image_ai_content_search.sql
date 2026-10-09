ALTER TABLE image_metadata
    ADD COLUMN IF NOT EXISTS ai_caption TEXT,
    ADD COLUMN IF NOT EXISTS ai_tags TEXT[],
    ADD COLUMN IF NOT EXISTS ai_ocr_text TEXT,
    ADD COLUMN IF NOT EXISTS ai_search_text TEXT;

CREATE INDEX IF NOT EXISTS idx_image_metadata_pending_ai_content
    ON image_metadata (file_id)
    WHERE ai_classified_at IS NOT NULL AND ai_caption IS NULL;

DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'pg_trgm extension unavailable, content search stays unindexed: %', SQLERRM;
    END;

    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        BEGIN
            CREATE INDEX IF NOT EXISTS idx_image_metadata_ai_search_text_trigram
                ON image_metadata USING gin (ai_search_text gin_trgm_ops)
                WHERE ai_search_text IS NOT NULL;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'ai_search_text trigram index not created: %', SQLERRM;
        END;
    END IF;
END
$$;
