DO $$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'pg_trgm extension unavailable, audio search stays unindexed: %', SQLERRM;
    END;

    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        BEGIN
            CREATE INDEX IF NOT EXISTS idx_audio_metadata_artist_trigram
                ON audio_metadata USING gin (lower(artist) gin_trgm_ops);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'artist trigram index not created: %', SQLERRM;
        END;

        BEGIN
            CREATE INDEX IF NOT EXISTS idx_audio_metadata_album_artist_trigram
                ON audio_metadata USING gin (lower(album_artist) gin_trgm_ops);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'album_artist trigram index not created: %', SQLERRM;
        END;

        BEGIN
            CREATE INDEX IF NOT EXISTS idx_audio_metadata_album_trigram
                ON audio_metadata USING gin (lower(album) gin_trgm_ops);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'album trigram index not created: %', SQLERRM;
        END;
    END IF;
END
$$;
