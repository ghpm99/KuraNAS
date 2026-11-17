CREATE TABLE IF NOT EXISTS audio_metadata_override (
    file_id INTEGER PRIMARY KEY REFERENCES home_file (id) ON DELETE CASCADE,
    title TEXT NULL,
    artist TEXT NULL,
    album TEXT NULL,
    album_artist TEXT NULL,
    genre TEXT NULL,
    year TEXT NULL,
    disc_number INTEGER NULL,
    track_number INTEGER NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE VIEW audio_metadata_effective AS
SELECT
    am.id,
    am.file_id,
    am.path,
    am.mime,
    am.length,
    am.bitrate,
    am.sample_rate,
    am.channels,
    am.bitrate_mode,
    am.encoder_info,
    am.bit_depth,
    COALESCE(o.title, am.title) AS title,
    COALESCE(o.artist, am.artist) AS artist,
    COALESCE(o.album, am.album) AS album,
    COALESCE(o.album_artist, am.album_artist) AS album_artist,
    COALESCE(o.track_number::TEXT, am.track_number) AS track_number,
    COALESCE(o.genre, am.genre) AS genre,
    am.composer,
    COALESCE(o.year, am.year) AS year,
    am.recording_date,
    am.encoder,
    am.publisher,
    am.original_release_date,
    am.original_artist,
    am.lyricist,
    am.lyrics,
    am.created_at,
    COALESCE(o.disc_number, am.disc_number) AS disc_number,
    am.disc_total,
    COALESCE(o.track_number, am.track_no) AS track_no,
    am.track_total,
    am.tags_extracted_version,
    am.catalog_artist_key,
    am.catalog_artist_label,
    am.catalog_album_key,
    am.catalog_album_label,
    am.catalog_genre_keys,
    am.catalog_genre_labels,
    am.catalog_album_base_key,
    am.catalog_album_artist_label
FROM
    audio_metadata am
    LEFT JOIN audio_metadata_override o ON o.file_id = am.file_id;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        RAISE NOTICE 'pg_trgm extension unavailable, audio override search stays unindexed';
        RETURN;
    END IF;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_override_title_fold_trigram
            ON audio_metadata_override USING gin (kuranas_fold(title) gin_trgm_ops)
            WHERE title IS NOT NULL;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'override title trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_override_artist_fold_trigram
            ON audio_metadata_override USING gin (kuranas_fold(artist) gin_trgm_ops)
            WHERE artist IS NOT NULL;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'override artist trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_override_album_fold_trigram
            ON audio_metadata_override USING gin (kuranas_fold(album) gin_trgm_ops)
            WHERE album IS NOT NULL;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'override album trigram index not created: %', SQLERRM;
    END;

    BEGIN
        CREATE INDEX IF NOT EXISTS idx_audio_metadata_override_album_artist_fold_trigram
            ON audio_metadata_override USING gin (kuranas_fold(album_artist) gin_trgm_ops)
            WHERE album_artist IS NOT NULL;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'override album_artist trigram index not created: %', SQLERRM;
    END;
END
$$;
