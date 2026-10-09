DO $$
DECLARE
    unaccent_schema TEXT;
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS unaccent;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'unaccent extension unavailable, search stays accent-sensitive: %', SQLERRM;
    END;

    SELECT namespace.nspname
    INTO unaccent_schema
    FROM pg_extension extension
        JOIN pg_namespace namespace ON namespace.oid = extension.extnamespace
    WHERE extension.extname = 'unaccent';

    IF unaccent_schema IS NOT NULL THEN
        BEGIN
            EXECUTE format(
                $fold$
                CREATE OR REPLACE FUNCTION kuranas_fold(text) RETURNS text
                    LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
                    AS 'SELECT lower(%1$I.unaccent(''%1$I.unaccent''::regdictionary, $1))'
                $fold$,
                unaccent_schema
            );
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'accent-folding function not created: %', SQLERRM;
            unaccent_schema := NULL;
        END;
    END IF;

    IF unaccent_schema IS NULL THEN
        CREATE OR REPLACE FUNCTION kuranas_fold(text) RETURNS text
            LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
            AS 'SELECT lower($1)';
    END IF;
END
$$;

CREATE OR REPLACE FUNCTION kuranas_fold_terms(text[]) RETURNS text[]
    LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
    AS 'SELECT ARRAY(SELECT kuranas_fold(search_term) FROM unnest($1) AS search_term)';

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
        CREATE INDEX IF NOT EXISTS home_file_active_name_fold_trigram
            ON home_file USING gin (kuranas_fold(name) gin_trgm_ops)
            WHERE deleted_at IS NULL;
        DROP INDEX IF EXISTS home_file_active_name_trigram;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'name fold trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_artist_fold_trigram
            ON audio_metadata USING gin (kuranas_fold(artist) gin_trgm_ops);
        DROP INDEX IF EXISTS idx_audio_metadata_artist_trigram;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'artist fold trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_album_artist_fold_trigram
            ON audio_metadata USING gin (kuranas_fold(album_artist) gin_trgm_ops);
        DROP INDEX IF EXISTS idx_audio_metadata_album_artist_trigram;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'album_artist fold trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_album_fold_trigram
            ON audio_metadata USING gin (kuranas_fold(album) gin_trgm_ops);
        DROP INDEX IF EXISTS idx_audio_metadata_album_trigram;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'album fold trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_title_fold_trigram
            ON audio_metadata USING gin (kuranas_fold(title) gin_trgm_ops);
        DROP INDEX IF EXISTS idx_audio_metadata_title_trigram;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'title fold trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_image_metadata_ai_search_text_fold_trigram
            ON image_metadata USING gin (kuranas_fold(ai_search_text) gin_trgm_ops)
            WHERE ai_search_text IS NOT NULL;
        DROP INDEX IF EXISTS idx_image_metadata_ai_search_text_trigram;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'ai_search_text fold trigram index not created: %', SQLERRM;
    END;
END
$$;
