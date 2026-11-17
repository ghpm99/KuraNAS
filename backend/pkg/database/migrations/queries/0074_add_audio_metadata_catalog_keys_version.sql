ALTER TABLE audio_metadata
    ADD COLUMN IF NOT EXISTS catalog_keys_version INTEGER NOT NULL DEFAULT 0;

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
    am.catalog_album_artist_label,
    am.catalog_keys_version
FROM
    audio_metadata am
    LEFT JOIN audio_metadata_override o ON o.file_id = am.file_id;
