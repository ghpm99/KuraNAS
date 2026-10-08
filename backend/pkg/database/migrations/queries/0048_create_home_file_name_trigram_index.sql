DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'pg_trgm extension unavailable, name search stays unindexed: %', SQLERRM;
    END;

    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        BEGIN
            CREATE INDEX IF NOT EXISTS home_file_active_name_trigram
                ON home_file USING gin (lower(name) gin_trgm_ops)
                WHERE deleted_at IS NULL;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'name trigram index not created: %', SQLERRM;
        END;
    END IF;
END
$$;
