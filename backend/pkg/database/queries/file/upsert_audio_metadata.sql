INSERT INTO
    audio_metadata (
        file_id,
        PATH,
        mime,
        LENGTH,
        bitrate,
        sample_rate,
        channels,
        bitrate_mode,
        encoder_info,
        bit_depth,
        title,
        artist,
        album,
        album_artist,
        track_number,
        genre,
        composer,
        YEAR,
        recording_date,
        encoder,
        publisher,
        original_release_date,
        original_artist,
        lyricist,
        lyrics,
        created_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26) ON CONFLICT (file_id, PATH)
DO
UPDATE
SET
    mime = EXCLUDED.mime,
    LENGTH = EXCLUDED.length,
    bitrate = EXCLUDED.bitrate,
    sample_rate = EXCLUDED.sample_rate,
    channels = EXCLUDED.channels,
    bitrate_mode = EXCLUDED.bitrate_mode,
    encoder_info = EXCLUDED.encoder_info,
    bit_depth = EXCLUDED.bit_depth,
    title = EXCLUDED.title,
    artist = EXCLUDED.artist,
    album = EXCLUDED.album,
    album_artist = EXCLUDED.album_artist,
    track_number = EXCLUDED.track_number,
    genre = EXCLUDED.genre,
    composer = EXCLUDED.composer,
    YEAR = EXCLUDED.year,
    recording_date = EXCLUDED.recording_date,
    encoder = EXCLUDED.encoder,
    publisher = EXCLUDED.publisher,
    original_release_date = EXCLUDED.original_release_date,
    original_artist = EXCLUDED.original_artist,
    lyricist = EXCLUDED.lyricist,
    lyrics = EXCLUDED.lyrics
RETURNING
    id,
    created_at;