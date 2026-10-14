DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'pg_trgm extension unavailable, folded search stays unindexed: %', SQLERRM;
    END;

    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        RETURN;
    END IF;

    BEGIN
        CREATE INDEX IF NOT EXISTS home_file_active_parent_path_fold_trigram
            ON home_file USING gin (kuranas_fold(parent_path) gin_trgm_ops)
            WHERE deleted_at IS NULL;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'parent_path fold trigram index not created: %', SQLERRM;
    END;
END
$$;
