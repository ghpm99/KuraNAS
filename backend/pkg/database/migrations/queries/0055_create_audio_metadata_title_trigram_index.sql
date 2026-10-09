DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'pg_trgm extension unavailable, audio title search stays unindexed: %', SQLERRM;
    END;

    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        BEGIN
            CREATE INDEX IF NOT EXISTS idx_audio_metadata_title_trigram
                ON audio_metadata USING gin (lower(title) gin_trgm_ops);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'title trigram index not created: %', SQLERRM;
        END;
    END IF;
END
$$;
